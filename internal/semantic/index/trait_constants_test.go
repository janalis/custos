package index

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestTraitConstantOwners(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "constants.php", `<?php
trait Inner { public const self|null VALUE = null; }
trait Outer { use Inner; }
trait Duplicate { public const self|null VALUE = null; }
trait Different { public const int VALUE = 1; }
trait Left { use Inner; }
trait Right { use Inner; }
class Root { public const string VALUE = 'root'; }
class Base extends Root { use Outer; }
class Child extends Base {}
class Other { use Outer; }
class Reimport extends Base { use Outer; }
class Diamond { use Left, Right; }
class Compatible { use Inner, Duplicate; }
class Conflict { use Inner, Different; }
class NestedConflict { use ConflictTrait; }
trait ConflictTrait { use Inner, Different; }
class Own { use Inner, Different; public const int VALUE = 2; }
interface Contract { public const int CONTRACT = 3; }
class ContractUser extends Root implements Contract { use Outer; }
`))
	for _, tc := range []struct{ class, owner string }{
		{"Base", "Base"},
		{"Child", "Base"},
		{"Other", "Other"},
		{"Reimport", "Reimport"},
		{"Diamond", "Diamond"},
		{"Compatible", "Compatible"},
		{"ContractUser", "ContractUser"},
	} {
		got := ix.FindConst(tc.class, "VALUE", 0)
		if got == nil || got.TypeClass != tc.owner {
			t.Fatalf("%s: %+v", tc.class, got)
		}
	}
	if ix.Class("Inner", 0).Consts["VALUE"].TypeClass != "" {
		t.Fatal("mutated trait declaration")
	}
	if got := ix.FindConst("Inner", "VALUE", 0); got == nil || got.TypeClass != "" {
		t.Fatalf("direct trait: %+v", got)
	}
	if got := ix.FindConst("Outer", "VALUE", 0); got == nil || got.TypeClass != "" {
		t.Fatalf("nested trait: %+v", got)
	}
	if got := ix.FindConst("Own", "VALUE", 0); got == nil || got.Class != "Own" || got.TypeClass != "" {
		t.Fatalf("own: %+v", got)
	}
	for _, class := range []string{"Conflict", "NestedConflict", "Unknown"} {
		if got := ix.FindConst(class, "VALUE", 0); got != nil {
			t.Fatalf("%s: %+v", class, got)
		}
	}
	if got := ix.FindConst("ContractUser", "CONTRACT", 0); got == nil || got.Class != "Contract" {
		t.Fatalf("interface: %+v", got)
	}
	if ix.FindConst("Child", "missing", 0) != nil {
		t.Fatal("missing constant")
	}
	// Direct lookup metadata must never become persistent stub metadata.
	imported := ix.FindConst("Base", "VALUE", 0)
	data, err := json.Marshal(imported)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "TypeClass") || strings.Contains(string(data), "Base") {
		t.Fatalf("transient owner encoded: %s", data)
	}
	var restored ClassConst
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.TypeClass != "" || restored.Type != imported.Type {
		t.Fatalf("round trip: %+v", restored)
	}
}

func TestConstantLookupBoundsAndPrecedence(t *testing.T) {
	ix := New(nil)
	classes := []*Class{
		{FQN: "Cycle", Kind: syntax.KindTrait, Traits: []string{"Cycle"}},
		{FQN: "CyclicClass", Traits: []string{"Cycle"}, Parent: "Present"},
		{FQN: "Missing", Traits: []string{"Absent"}},
		{FQN: "Interfaces", Parent: "Absent", Interfaces: []string{"Absent", "Present"}},
		{FQN: "Present", Consts: map[string]*ClassConst{"K": {Name: "K", Class: "Present", Value: "1"}}},
		{FQN: "DeepTrait", Kind: syntax.KindTrait, Traits: []string{"Leaf"}},
		{FQN: "Leaf", Kind: syntax.KindTrait, Consts: map[string]*ClassConst{"K": {Name: "K", Class: "Leaf", Value: "2"}}},
		{FQN: "Parent", Traits: []string{"DeepTrait"}},
		{FQN: "Child", Parent: "Parent", Interfaces: []string{"Present"}},
		{FQN: "Versioned", Kind: syntax.KindTrait, Avail: Avail{To: phpversion.PHP83}, Consts: map[string]*ClassConst{"K": {Name: "K", Class: "Versioned", Value: "83"}}},
		{FQN: "Versioned", Kind: syntax.KindTrait, Avail: Avail{From: phpversion.PHP84}, Consts: map[string]*ClassConst{"K": {Name: "K", Class: "Versioned", Value: "84"}}},
		{FQN: "VersionUser", Traits: []string{"Versioned"}},
	}
	for i := range MaxAncestors + 1 {
		classes = append(classes, &Class{FQN: fmt.Sprint(i), Kind: syntax.KindTrait, Traits: []string{fmt.Sprint(i + 1)}})
	}
	ix.Add(&FileSymbols{Path: "bounds.php", Classes: classes})
	for _, class := range []string{"Cycle", "CyclicClass", "Missing", "0", "Absent"} {
		if got := ix.FindConst(class, "K", 0); got != nil {
			t.Fatalf("%s: %+v", class, got)
		}
	}
	if got := ix.FindConst("Interfaces", "K", 0); got == nil || got.Class != "Present" {
		t.Fatalf("interface: %+v", got)
	}
	if got := ix.FindConst("Child", "K", 0); got == nil || got.Class != "Leaf" || got.TypeClass != "Parent" {
		t.Fatalf("parent before interface: %+v", got)
	}
	if ix.FindConst("VersionUser", "K", phpversion.PHP83).Value != "83" || ix.FindConst("VersionUser", "K", phpversion.PHP84).Value != "84" {
		t.Fatal("version availability")
	}
}

func TestCompatibleConstants(t *testing.T) {
	a := ClassConst{Type: "int", Value: "1"}
	for _, mutate := range []func(*ClassConst){
		func(c *ClassConst) { c.Type = "string" }, func(c *ClassConst) { c.Value = "2" }, func(c *ClassConst) { c.Visibility = Private }, func(c *ClassConst) { c.Final = true }, func(c *ClassConst) { c.Case = true },
	} {
		b := a
		mutate(&b)
		if compatibleConstants(&a, &b) {
			t.Fatalf("compatible: %+v", b)
		}
	}
}

func BenchmarkFindConstRelative(b *testing.B) {
	ix := New(nil)
	ix.Add(Extract(syntax.Parse("bench.php", []byte(`<?php trait T { public const self|null K = null; } class Plain { public const int K = 1; } class Base { use T; } class Child extends Base {} class Ordinary extends Plain {}`), syntax.Options{})))
	for _, class := range []string{"Plain", "Ordinary", "Child"} {
		b.Run(class, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				ix.FindConst(class, "K", 0)
			}
		})
	}
}
