package analysis

import (
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/names"
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

// namesNopRule reports each `;` after resolving names, which the test
// makes crash: its checks are skipped, the rule itself is not disabled.
type namesNopRule struct{ crashRule }

func (namesNopRule) Check(ctx *Context, n syntax.Node) {
	ctx.Names()
	ctx.ReportNode(n, "nop")
}

// TestRuleEnabledAfterSkippedCheck: a rule a check of which was skipped
// for lack of shared data no longer counts as enabled, whichever of the
// two runs first.
func TestRuleEnabledAfterSkippedCheck(t *testing.T) {
	old := newResolver
	t.Cleanup(func() { newResolver = old })
	newResolver = func(*syntax.File) *names.Resolver { panic("resolver bug") }
	for _, rules := range [][]Rule{{namesNopRule{}, deferRule{}}, {deferRule{}, namesNopRule{}}} {
		e, err := NewEngine(rules, Config{EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			src  string
			want []string
		}{
			{"<?php ;\n$a = !$b;", []string{"internal", "NestedNotOperators"}},
			// No check of the rule was skipped: it still runs.
			{"<?php f();\n$a = !$b;", nil},
			// Nothing defers to it: one run is enough.
			{"<?php ;", []string{"internal"}},
		} {
			var got []string
			for _, fd := range e.Analyze(syntax.Parse("x.php", []byte(tc.src), syntax.Options{Version: phpversion.PHP84})) {
				got = append(got, fd.Rule)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("%q: got %v, want %v", tc.src, got, tc.want)
			}
		}
	}
}
