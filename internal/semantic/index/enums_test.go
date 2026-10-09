package index

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEnumRuntimeMembers(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "interfaces.php", `<?php interface UnitEnum {} interface BackedEnum extends UnitEnum {}`))
	ix.Add(extract(t, "enums.php", `<?php
namespace App;
enum Flag { case Ready; }
enum Code: int { case Ready = 1; }
enum Label: string { case Ready = 'ready'; }
`))
	for _, name := range []string{"Flag", "Code", "Label"} {
		fqn := `App\` + name
		c := ix.Class(fqn, phpversion.PHP81)
		if c == nil || !ix.IsSubtype(fqn, "UnitEnum", phpversion.PHP81) || c.Avail.In(phpversion.PHP80) {
			t.Fatalf("enum availability/interfaces: %+v", c)
		}
		if p := ix.FindProperty(fqn, "name", phpversion.PHP81); p == nil || !p.Readonly || p.Type != "string" {
			t.Fatalf("name: %+v", p)
		}
		if m := ix.FindMethod(fqn, "cases", phpversion.PHP81); m == nil || !m.Static || !m.Builtin || m.DocReturn != `list<\`+fqn+`>` {
			t.Fatalf("cases: %+v", m)
		}
		if name == "Flag" {
			if ix.FindProperty(fqn, "value", phpversion.PHP81) != nil || ix.FindMethod(fqn, "from", phpversion.PHP81) != nil || ix.IsSubtype(fqn, "BackedEnum", phpversion.PHP81) {
				t.Fatal("unit enum has backed members")
			}
			continue
		}
		backing := "int"
		if name == "Label" {
			backing = "string"
		}
		if p := ix.FindProperty(fqn, "value", phpversion.PHP81); p == nil || !p.Readonly || p.Type != backing || !ix.IsSubtype(fqn, "BackedEnum", phpversion.PHP81) {
			t.Fatalf("value: %+v", p)
		}
		for _, method := range []string{"from", "tryfrom"} {
			m := ix.FindMethod(fqn, method, phpversion.PHP81)
			if m == nil || len(m.Params) != 1 || m.Params[0].Name != "value" || m.Params[0].Type != "int|string" {
				t.Fatalf("%s: %+v", method, m)
			}
		}
	}
}

func TestEnumPreservesExplicitMethodsAndAvailability(t *testing.T) {
	c := &Class{FQN: "E", Methods: map[string]*Method{"cases": {Name: "cases", Return: "string"}}, Props: map[string]*Property{}, Avail: Avail{From: phpversion.PHP85}}
	addEnumMembers(c, "")
	if c.Methods["cases"].Return != "string" || c.Avail.From != phpversion.PHP85 {
		t.Fatal("explicit declaration overwritten")
	}
}

func TestDeclaredClassConstantTypes(t *testing.T) {
	fs := extract(t, "constants.php", `<?php namespace App; use Vendor\Thing as Item;
class C { public const int COUNT = SOME_VALUE; public const Item|null ITEM = null; public const LEGACY = 1; }
`)
	c := fs.Classes[0]
	if c.Consts["COUNT"].Type != "int" || c.Consts["ITEM"].Type != `\Vendor\Thing|null` || c.Consts["LEGACY"].Type != "" {
		t.Fatalf("types: %+v %+v %+v", c.Consts["COUNT"], c.Consts["ITEM"], c.Consts["LEGACY"])
	}
}

func BenchmarkEnumExtraction(b *testing.B) {
	src := []byte(`<?php namespace App; enum Code: int { case Ready = 1; }`)
	b.ReportAllocs()
	for b.Loop() {
		Extract(syntax.Parse("bench.php", src, syntax.Options{Version: phpversion.PHP85}))
	}
}
