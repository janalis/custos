package index

import (
	"fmt"
	"testing"

	"custos/internal/syntax"
)

func TestRelativeTraitOwners(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "relative.php", `<?php
/** @method self magic() */ trait Inner { public self $p; public function m(): self {} }
trait Outer { use Inner { m as alias; } }
class Root {}
class Base extends Root { use Outer; }
class Child extends Base {}
class Reimport extends Base { use Outer; }
interface Hook { public string $hook { get; } }
class HookUser implements Hook {}
`))
	for _, tc := range []struct{ receiver, owner string }{{"Base", "Base"}, {"Child", "Base"}, {"Reimport", "Reimport"}} {
		for _, name := range []string{"m", "alias", "magic"} {
			m := ix.FindMethod(tc.receiver, name, 0)
			if m == nil || m.TypeClass != tc.owner || m.Class != "Inner" {
				t.Fatalf("%s %s: %+v", tc.receiver, name, m)
			}
		}
		p := ix.FindProperty(tc.receiver, "p", 0)
		if p == nil || p.TypeClass != tc.owner || p.Class != "Inner" {
			t.Fatalf("%s: %+v", tc.receiver, p)
		}
	}
	if ix.Class("Inner", 0).Methods["m"].TypeClass != "" || ix.Class("Inner", 0).Props["p"].TypeClass != "" {
		t.Fatal("mutated declaration")
	}
	if ix.FindProperty("HookUser", "hook", 0) == nil {
		t.Fatal("interface property")
	}
	if p := ix.FindProperty("Inner", "p", 0); p.TypeClass != "" {
		t.Fatal("trait must stay unbound")
	}
	if ix.FindProperty("Child", "missing", 0) != nil {
		t.Fatal("missing property")
	}
	if ix.FindProperty("absent", "missing", 0) != nil {
		t.Fatal("missing class")
	}
}

func TestPropertyLookupBounds(t *testing.T) {
	ix := New(nil)
	classes := []*Class{{FQN: "Cycle", Kind: syntax.KindTrait, Traits: []string{"Cycle"}}, {FQN: "Missing", Traits: []string{"Absent"}}, {FQN: "Interfaces", Interfaces: []string{"Absent", "Present"}}, {FQN: "Present", Props: map[string]*Property{"p": {Name: "p"}}}}
	for i := range MaxAncestors + 1 {
		classes = append(classes, &Class{FQN: fmt.Sprint(i), Kind: syntax.KindTrait, Traits: []string{fmt.Sprint(i + 1)}})
	}
	ix.Add(&FileSymbols{Path: "bound.php", Classes: classes})
	for _, class := range []string{"Cycle", "Missing", "0", "Absent"} {
		l := propertyLookup{ix: ix}
		if l.find(class, "p", "") != nil {
			t.Fatal(class)
		}
	}
	l := propertyLookup{ix: ix}
	if l.find("Interfaces", "p", "") == nil {
		t.Fatal("interface fallback")
	}
}

func BenchmarkFindPropertyRelative(b *testing.B) {
	ix := New(nil)
	ix.Add(Extract(syntax.Parse("bench.php", []byte(`<?php trait T { public self $p; } class Plain { public int $p; } class Base { use T; } class Child extends Base {} class Ordinary extends Plain {}`), syntax.Options{})))
	for _, class := range []string{"Plain", "Ordinary", "Child"} {
		b.Run(class, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				ix.FindProperty(class, "p", 0)
			}
		})
	}
}
