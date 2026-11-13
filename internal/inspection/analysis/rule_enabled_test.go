package analysis

import (
	"slices"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/names"
)

// deferRule reports each `!` unless UnnecessarySemicolon (crashRule) runs.
type deferRule struct{ okRule }

func (deferRule) Check(ctx *Context, n syntax.Node) {
	if !ctx.RuleChecks("UnnecessarySemicolon", n) {
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

// namesFileRule (UnnecessarySemicolon) reports the file; it resolves
// names first when the file mentions $skip, which the test makes crash:
// its check is skipped, the rule itself is not disabled.
type namesFileRule struct{ patternRule }

func (namesFileRule) CheckFile(ctx *Context) {
	if strings.Contains(string(ctx.Src), "$skip") {
		ctx.Names()
	}
	ctx.ReportNode(ctx.File, "nop")
}

// TestRuleChecksAfterSkippedFileCheck: a file rule whose check was
// skipped for lack of shared data no longer checks any node, whichever of
// the two runs first.
func TestRuleChecksAfterSkippedFileCheck(t *testing.T) {
	old := newResolver
	t.Cleanup(func() { newResolver = old })
	newResolver = func(*syntax.File) *names.Resolver { panic("resolver bug") }
	owner := namesFileRule{patternRule{"UnnecessarySemicolon"}}
	for _, rules := range [][]Rule{{owner, deferRule{}}, {deferRule{}, owner}} {
		e, err := NewEngine(rules, Config{EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			src  string
			want []string
		}{
			{"<?php $skip;\n$a = !$b;", []string{"internal", "NestedNotOperators"}},
			// Its check was not skipped: it still runs.
			{"<?php f();\n$a = !$b;", []string{"UnnecessarySemicolon"}},
			// Nothing defers to it: one run is enough.
			{"<?php $skip;", []string{"internal"}},
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

// ownerRule (UnnecessarySemicolon) reports each `!`; on `!$skip` its check
// resolves names first, which the test makes crash.
type ownerRule struct{ okRule }

func (ownerRule) ID() string { return "UnnecessarySemicolon" }

func (ownerRule) Check(ctx *Context, n syntax.Node) {
	if v, ok := n.(*syntax.Unary).Expr.(*syntax.Variable); ok && v.Name == "skip" {
		ctx.Names()
	}
	ctx.ReportNode(n, "owner")
}

// nodeDeferRule reports each `!` the owner's check did not run on.
type nodeDeferRule struct{ okRule }

func (nodeDeferRule) Check(ctx *Context, n syntax.Node) {
	if !ctx.RuleChecks("UnnecessarySemicolon", n) {
		ctx.ReportNode(n, "deferred")
	}
}

// TestRuleCheckRanPerNode: a check skipped on one node leaves only that
// node to the rules deferring to its rule, whichever runs first.
func TestRuleCheckRanPerNode(t *testing.T) {
	old := newResolver
	t.Cleanup(func() { newResolver = old })
	newResolver = func(*syntax.File) *names.Resolver { panic("resolver bug") }
	for _, rules := range [][]Rule{{ownerRule{}, nodeDeferRule{}}, {nodeDeferRule{}, ownerRule{}}} {
		e, err := NewEngine(rules, Config{EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, fd := range e.Analyze(syntax.Parse("x.php", []byte("<?php $a = !$b;\n$c = !$skip;"), syntax.Options{Version: phpversion.PHP84})) {
			got = append(got, fd.Rule+":"+fd.Message)
		}
		want := []string{"UnnecessarySemicolon:owner", "NestedNotOperators:deferred"}
		if len(got) != 3 || !strings.HasPrefix(got[0], "internal:") || !slices.Equal(got[1:], want) {
			t.Errorf("got %v, want internal + %v", got, want)
		}
	}
}
