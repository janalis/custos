// Package ruletest runs one inspection over PHP snippets, for the tests of
// inspection packages that check cases too small for a fixture file.
package ruletest

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Case is a PHP snippet, without its opening tag, and how many findings
// the inspection must report on it.
type Case struct {
	Src  string
	Want int
}

// Count analyzes each case with rule alone, at the default PHP version, and
// requires Want findings, each with message over a nonempty range.
func Count(t testing.TB, rule analysis.Rule, message string, cases []Case) {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{rule}, analysis.Config{Only: []string{rule.ID()}})
	if err != nil {
		t.Fatal(err)
	} else {
		for _, tc := range cases {
			check(t, e, message, tc)
		}
	}
}

// check analyzes one case (a failure of t may not stop the caller).
func check(t testing.TB, e *analysis.Engine, message string, tc Case) {
	t.Helper()
	got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.Src), syntax.Options{}))
	if len(got) != tc.Want {
		t.Fatalf("%s: got %d findings, want %d: %+v", tc.Src, len(got), tc.Want, got)
	} else {
		for _, g := range got {
			if g.Message != message || g.Span.End <= g.Span.Start {
				t.Fatalf("%s: invalid finding: %+v", tc.Src, g)
			}
		}
	}
}
