package index

import (
	"encoding/json"
	"fmt"
	"testing"

	"custos/internal/phpver"
)

func TestTraitMethodComposition(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "traits.php", `<?php
 namespace Demo;
 trait A { public function run(): int {} public function solo(): string {} }
 trait B { public function run(): string {} }
 trait Nested { use B, A { A::run insteadof B; B::run as protected other; solo as private hidden; A::solo as public visible; A::solo as final sealed; } }
 class Base { public function inherited(): bool {} public function run(): bool {} }
 class C extends Base { use Nested; }
 class Override extends C { public function run(): float {} }
 class Conflict extends Base { use A, B; }
 class Reverse { use B, A { A::run insteadof B; } }
 interface Contract { public function contract(): int; }
 class Implements implements Contract {}
 /** @method string inherited() */ class Magic extends Base {}
 `))
	for _, tc := range []struct {
		class, name, ret string
		vis              Visibility
		final            bool
	}{
		{"C", "run", "int", Public, false},
		{"C", "other", "string", Protected, false},
		{"C", "hidden", "string", Private, false},
		{"C", "visible", "string", Public, false},
		{"C", "sealed", "string", Public, true},
		{"C", "solo", "string", Public, false},
		{"C", "inherited", "bool", Public, false},
		{"Override", "run", "float", Public, false},
		{"Reverse", "RUN", "int", Public, false},
		{"Implements", "contract", "int", Public, false},
		{"Magic", "inherited", "bool", Public, false},
	} {
		t.Run(tc.class+tc.name, func(t *testing.T) {
			m := ix.FindMethod("Demo\\"+tc.class, tc.name, phpver.PHP84)
			if m == nil || m.Return != tc.ret || m.Visibility != tc.vis || m.Final != tc.final {
				t.Fatalf("method: %+v", m)
			}
		})
	}
	if ix.FindMethod("Demo\\Conflict", "run", 0) != nil {
		t.Fatal("conflict must not inherit parent")
	}
	if ix.FindMethod("Demo\\C", "absent", 0) != nil {
		t.Fatal("missing")
	}
	a := ix.Class("Demo\\A", 0).Methods["solo"]
	if a.Visibility != Public || a.Final || a.Name != "solo" {
		t.Fatal("adaptation mutated declaration")
	}
	if m := ix.FindMethod("Demo\\C", "hidden", 0); m.Class != "Demo\\A" {
		t.Fatal("lost source")
	}
	if len(ix.Class("Demo\\Nested", 0).TraitAdaptations) != 5 {
		t.Fatal("extraction")
	}
	encoded, err := json.Marshal(ix.Class("Demo\\Nested", 0))
	if err != nil {
		t.Fatal(err)
	}
	var restored Class
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if len(restored.TraitAdaptations) != 5 || !restored.TraitAdaptations[4].Final || restored.TraitAdaptations[0].Trait != "Demo\\A" || len(restored.TraitAdaptations[0].Insteadof) != 1 {
		t.Fatalf("trait metadata round trip: %+v", restored.TraitAdaptations)
	}
}

func TestTraitMethodCyclesAndLimits(t *testing.T) {
	ix := New(nil)
	classes := []*Class{{FQN: "Cycle", Traits: []string{"Cycle"}}, {FQN: "Missing", Traits: []string{"Unknown"}}}
	for i := 0; i < MaxAncestors+1; i++ {
		classes = append(classes, &Class{FQN: fmt.Sprint(i), Traits: []string{fmt.Sprint(i + 1)}})
	}
	ix.Add(&FileSymbols{Path: "limits.php", Classes: classes})
	for _, c := range []string{"Cycle", "Missing", "0", "Unknown"} {
		if ix.FindMethod(c, "missing", 0) != nil {
			t.Fatal(c)
		}
	}
}

func BenchmarkFindMethodTrait(b *testing.B) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "bench.php", Classes: []*Class{
		{FQN: "t", Methods: map[string]*Method{"run": {Name: "run", Class: "T", Return: "int"}}},
		{FQN: "c", Traits: []string{"t"}},
	}})
	b.ReportAllocs()
	for b.Loop() {
		if ix.FindMethod("c", "run", 0) == nil {
			b.Fatal("missing")
		}
	}
}

func TestTraitConflictsAndDiamonds(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "diamond.php", `<?php
 trait RootTrait { function f(): int {} }
 trait LeftTrait { use RootTrait; }
 trait RightTrait { use RootTrait; }
 trait Changed { use RootTrait { f as private; } }
 trait OtherTrait { function f(): string {} }
 trait Bad { use RootTrait, OtherTrait; }
 class Diamond { use LeftTrait, RightTrait; }
 class ChangedDiamond { use LeftTrait, Changed; }
 class NestedBad { use Bad; }
 class AliasBad { use Bad { Bad::f as alias; } }
 class Ambiguous { use RootTrait, OtherTrait { f as alias; } }
 class DuplicateAlias { use RootTrait, OtherTrait { RootTrait::f as alias; OtherTrait::f as alias; } }
 class MissingAlias { use RootTrait { absent as alias; } }
 `))
	if m := ix.FindMethod("Diamond", "f", 0); m == nil || m.Class != "RootTrait" {
		t.Fatalf("diamond: %+v", m)
	}
	for _, tc := range []struct{ class, name string }{{"ChangedDiamond", "f"}, {"NestedBad", "f"}, {"AliasBad", "alias"}, {"Ambiguous", "alias"}, {"DuplicateAlias", "alias"}, {"MissingAlias", "alias"}} {
		if m := ix.FindMethod(tc.class, tc.name, 0); m != nil {
			t.Fatalf("%s: %+v", tc.class, m)
		}
	}
}

func BenchmarkFindMethodDirect(b *testing.B) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "direct.php", Classes: []*Class{{FQN: "c", Methods: map[string]*Method{"run": {Name: "run"}}}}})
	b.Run("composition", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			ix.FindMethod("c", "run", 0)
		}
	})
	b.Run("ancestors-baseline", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			for _, c := range ix.Ancestors("c", 0) {
				if c.Methods["run"] != nil {
					break
				}
			}
		}
	})
}

func TestAbstractTraitRequirements(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "requirements.php", `<?php
 trait Requirement { abstract public function f(): object; }
 trait AnotherRequirement { abstract public function f(): stdClass; }
 trait Implementation { public function f(): stdClass { return new stdClass; } }
 trait NestedRequirement { use Requirement; }
 class Forward { use Requirement, Implementation; }
 class Reverse { use Implementation, Requirement; }
 class Several { use Requirement, AnotherRequirement, Implementation; }
 class Base { public function f(): stdClass { return new stdClass; } public function alias(): stdClass { return new stdClass; } }
 class Inherited extends Base { use Requirement; }
 class Nested extends Base { use NestedRequirement; }
 class Aliased extends Base { use Requirement { Requirement::f as alias; } }
 abstract class Unsatisfied { use Requirement, AnotherRequirement; }
 abstract class ContractOnly { use Requirement; }
 trait SameRequirement { abstract public function f(): object; }
 abstract class Compatible { use Requirement, SameRequirement; }
 abstract class AbstractParent { abstract public function f(): object; }
 abstract class InheritContract extends AbstractParent {}
 interface AbstractInterface { public function f(): object; }
 abstract class InterfaceContract implements AbstractInterface {}
 abstract class CompatibleAlias { use Requirement, SameRequirement { f as alias; } }
 abstract class IncompatibleAlias { use Requirement, AnotherRequirement { f as alias; } }
 abstract class ParentWins extends Base { use Requirement, AnotherRequirement; }
 class UnqualifiedAlias { use Requirement, Implementation { f as alias; } }
 `))
	for _, tc := range []struct{ class, name, source string }{
		{"Forward", "f", "Implementation"},
		{"Reverse", "f", "Implementation"},
		{"Several", "f", "Implementation"},
		{"Inherited", "f", "Base"},
		{"Nested", "f", "Base"},
		{"Aliased", "alias", "Base"},
		{"UnqualifiedAlias", "alias", "Implementation"},
		{"ParentWins", "f", "Base"},
	} {
		m := ix.FindMethod(tc.class, tc.name, 0)
		if m == nil || m.Abstract || m.Class != tc.source || m.Return != "\\stdClass" {
			t.Fatalf("%s: %+v", tc.class, m)
		}
	}
	if ix.FindMethod("Unsatisfied", "f", 0) != nil {
		t.Fatal("requirements do not supply an implementation")
	}
	for _, tc := range []struct{ class, name string }{{"ContractOnly", "f"}, {"Compatible", "f"}, {"InheritContract", "f"}, {"InterfaceContract", "f"}, {"CompatibleAlias", "alias"}} {
		m := ix.FindMethod(tc.class, tc.name, 0)
		if m == nil || !m.Abstract || m.Return != "object" {
			t.Fatalf("contract %s: %+v", tc.class, m)
		}
	}
	if ix.FindMethod("IncompatibleAlias", "alias", 0) != nil {
		t.Fatal("incompatible alias requirements")
	}
	requirement := ix.Class("Requirement", 0).Methods["f"]
	if !requirement.Abstract || requirement.Name != "f" {
		t.Fatal("mutated requirement")
	}
}

func TestAbstractContractParameters(t *testing.T) {
	for _, tc := range []struct {
		a, b  *Method
		equal bool
	}{
		{&Method{Return: "object"}, &Method{Return: "object"}, true},
		{&Method{Params: []Param{{Name: "x", Type: "int"}}}, &Method{Params: []Param{{Name: "x", Type: "int"}}}, true},
		{&Method{Params: []Param{{Name: "x", Type: "int"}}}, &Method{Params: []Param{{Name: "x", Type: "string"}}}, false},
		{&Method{Params: []Param{{Name: "x"}}}, &Method{}, false},
	} {
		if sameAbstractContract(tc.a, tc.b) != tc.equal {
			t.Fatal("contract parameters")
		}
	}
}

// Incomplete project symbols can describe a concrete interface method while
// edits are in progress. Lookup still follows the indexed declaration.
func TestConcreteInterfaceMethodLookup(t *testing.T) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "editing.php", Classes: []*Class{
		{FQN: "contract", Methods: map[string]*Method{"f": {Name: "f", Return: "int"}}},
		{FQN: "receiver", Interfaces: []string{"contract"}},
	}})
	if m := ix.FindMethod("receiver", "f", 0); m == nil || m.Return != "int" {
		t.Fatal("indexed interface method")
	}
}

func TestVersionedTraitAdaptationCache(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "versioned-trait.php", `<?php
 trait VersionedTrait {
  #[LanguageLevelTypeAware(["8.1" => "mixed"], default: "array|false")]
  public function get() {}
 }
 class AliasedVersioned { use VersionedTrait { get as private previous; get as protected; } }
 `))
	original := ix.Class("VersionedTrait", 0).Methods["get"]
	// Populate the stable declaration/version entry before measuring repeated
	// adaptations. Transient aliases must never become cache keys.
	original.at(phpver.PHP74)
	size := func() int { n := 0; verCopies.Range(func(_, _ any) bool { n++; return true }); return n }
	before := size()
	for range 20 {
		alias := ix.FindMethod("AliasedVersioned", "previous", phpver.PHP74)
		adapted := ix.FindMethod("AliasedVersioned", "get", phpver.PHP74)
		if alias == nil || alias.Return != "array|false" || alias.Visibility != Private || alias.Name != "previous" || alias.Class != "VersionedTrait" {
			t.Fatalf("alias: %+v", alias)
		}
		if adapted == nil || adapted.Return != "array|false" || adapted.Visibility != Protected {
			t.Fatalf("adapted: %+v", adapted)
		}
	}
	if after := size(); after != before {
		t.Fatalf("transient version copies retained: %d -> %d", before, after)
	}
	if original.Return != "mixed" || original.Visibility != Public || original.Name != "get" || len(original.RetVer) == 0 {
		t.Fatal("mutated original versioned method")
	}
	newest := ix.FindMethod("AliasedVersioned", "previous", phpver.PHP84)
	if newest.Return != "mixed" {
		t.Fatalf("new version: %+v", newest)
	}
}
