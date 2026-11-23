package pattern

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// TestPatternsSeePartialApplications: a pattern rule sees first-class
// callables and partial applications as written, so a ban on a function
// cannot be bypassed by making a closure of it.
func TestPatternsSeePartialApplications(t *testing.T) {
	r, err := Compile(Definition{ID: "NoUnserialize", Message: "m", Match: Patterns{`\unserialize(...)`}})
	if err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{PHP: phpversion.PHP86})
	if err != nil {
		t.Fatal(err)
	}
	src := "<?php $f = unserialize(...); $g = array_map(unserialize(...), $_POST); $h = unserialize(?, []); $i = unserialize($s);"
	if got := len(e.Analyze(syntax.Parse("p.php", []byte(src), syntax.Options{Version: phpversion.PHP86}))); got != 4 {
		t.Fatalf("findings: %d, want 4", got)
	}
}

// TestPatternWildcardsNeedNoClosure: `...` stays a wildcard where PHP could
// not build a closure (`new C(...)`, `$o?->m(...)`).
func TestPatternWildcardsNeedNoClosure(t *testing.T) {
	for _, match := range []string{`new \DateTime(...)`, `$o?->m(...)`} {
		if _, err := Compile(Definition{ID: "R", Message: "m", Match: Patterns{match}}); err != nil {
			t.Errorf("%s: %v", match, err)
		}
	}
	if _, err := Compile(Definition{ID: "R", Message: "m", Match: Patterns{`f(,`}}); err == nil {
		t.Fatal("a pattern that does not parse is refused")
	}
}

// TestWildcardMatchesPartialApplications: `f(...)` in a pattern matches a
// partial application of f; a metavariable does not match a `?`.
func TestWildcardMatchesPartialApplications(t *testing.T) {
	src := "<?php $a = dump(?, 1); $b = dump(...);"
	for match, want := range map[string]int{`\dump(...)`: 2, `\dump($X, 1)`: 0} {
		r, err := Compile(Definition{ID: "R", Message: "m", Match: Patterns{match}})
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{PHP: phpversion.PHP86})
		if err != nil {
			t.Fatal(err)
		}
		if got := len(e.Analyze(syntax.Parse("p.php", []byte(src), syntax.Options{Version: phpversion.PHP86}))); got != want {
			t.Errorf("%s: %d findings, want %d", match, got, want)
		}
	}
}
