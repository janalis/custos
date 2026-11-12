package analysis

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// deferRule reports each `!` unless UnnecessarySemicolon (crashRule) runs.
type deferRule struct{ okRule }

func (deferRule) Check(ctx *Context, n syntax.Node) {
	if !ctx.RuleEnabled("UnnecessarySemicolon") {
		ctx.ReportNode(n, "found")
	}
}

// TestRuleEnabledAfterCrash: a rule skipped on a file after crashing there
// no longer counts as enabled, so what others leave to it is reported.
func TestRuleEnabledAfterCrash(t *testing.T) {
	f := func(src string) []string {
		e, err := NewEngine([]Rule{crashRule{}, deferRule{}}, Config{EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		var rules []string
		for _, fd := range e.Analyze(syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpversion.PHP84})) {
			rules = append(rules, fd.Rule)
		}
		return rules
	}
	if got := f("<?php $a = !$b;"); len(got) != 0 {
		t.Errorf("deferred to a running rule: %v", got)
	}
	if got := f("<?php ;\n$a = !$b;"); len(got) != 2 || got[0] != "internal" || got[1] != "NestedNotOperators" {
		t.Errorf("deferred to a crashed rule: %v", got)
	}
}

// pnaRule reports each `!` as PossiblyNullAccess, which the suppressions of
// NullPointerException silence too.
type pnaRule struct{ okRule }

func (pnaRule) ID() string { return "PossiblyNullAccess" }

func TestSuppressedByOtherRule(t *testing.T) {
	e, err := NewEngine([]Rule{pnaRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nf();\n$a = !$b;", 1},
		{"<?php\nf();\n// @noinspection NullPointerException\n$a = !$b;", 0},
		{"<?php\nf();\n/** @noinspection NullPointerExceptionInspection */\n$a = !$b;", 0},
		{"<?php\nf();\n// @noinspection PossiblyNullArgument\n$a = !$b;", 1},
	} {
		if got := e.Analyze(syntax.Parse("x.php", []byte(tc.src), syntax.Options{Version: phpversion.PHP84})); len(got) != tc.want {
			t.Errorf("%q: got %+v, want %d findings", tc.src, got, tc.want)
		}
	}
}
