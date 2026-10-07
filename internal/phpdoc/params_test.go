package phpdoc

import (
	"strings"
	"testing"
)

func TestParamsNameFirst(t *testing.T) {
	d := Parse(`/**
 * @param int $a first
 * @param $b stdClass
 * @param $c string|null the c
 * @param $d the value
 * @param $e
 */`)
	want := []Param{{"int", "a"}, {"stdClass", "b"}, {"string|null", "c"}, {"", "d"}, {"", "e"}}
	got := d.Params()
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("param %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestTypeAliasCap(t *testing.T) {
	d := Parse("/**\n * @phpstan-type Big array{k: " + strings.Repeat("x", MaxAliasLen) + "}\n * @phpstan-type Small int\n */")
	a := d.TypeAliases()
	if a["Big"] != "" || a["Small"] != "int" {
		t.Errorf("got %q / %q", a["Big"][:min(len(a["Big"]), 20)], a["Small"])
	}
	if _, ok := a["Big"]; !ok {
		t.Error("an over-long alias must still be declared (as mixed)")
	}
}

func TestVarTypeNameFirstOtherVariable(t *testing.T) {
	d := Parse("/**\n * @var $a int\n * @var $b string\n */")
	if got := d.VarType("b"); got != "string" {
		t.Errorf("VarType(b) = %q", got)
	}
	if got := d.VarType("c"); got != "" {
		t.Errorf("VarType(c) = %q, want none", got)
	}
}

func TestEmptyTemplateAndImportTags(t *testing.T) {
	d := Parse("/**\n * @template\n * @template T of Foo\n * @phpstan-import-type\n * @phpstan-import-type A from B as C\n */")
	if tp := d.TemplateParams(); len(tp) != 1 || tp[0] != (TemplateParam{"T", "Foo"}) {
		t.Errorf("TemplateParams = %+v", tp)
	}
	if a := d.TypeAliases(); len(a) != 1 || a["C"] != "" {
		t.Errorf("TypeAliases = %v", a)
	}
}

func TestTagMissing(t *testing.T) {
	d := Parse("/** @return int */")
	if d.Has("param") || !d.Has("return") {
		t.Error("Has")
	}
}
