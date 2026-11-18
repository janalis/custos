package analysis

import (
	"reflect"
	"runtime/pprof"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestLabelRules(t *testing.T) {
	rules := []Rule{plainFileRule{patternRule{"UnnecessarySemicolon"}}, okRule{}}
	file := func() *syntax.File {
		return syntax.Parse("x.php", []byte("<?php $a = !$b;"), syntax.Options{Version: phpversion.PHP84})
	}
	plain, err := NewEngine(rules, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	if plain.labels != nil {
		t.Fatal("checks are labeled only when asked")
	}
	labeled, err := NewEngine(rules, Config{EnableAll: true, LabelRules: true})
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range labeled.Rules() {
		if got, _ := pprof.Label(labeled.labels[i], "rule"); got != id {
			t.Fatalf("rule %d labeled %q, want %q", i, got, id)
		}
	}
	want, got := plain.Analyze(file()), labeled.Analyze(file())
	if len(got) != 2 || !reflect.DeepEqual(got, want) {
		t.Fatalf("labeling changed findings: %+v, want %+v", got, want)
	}
}
