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
