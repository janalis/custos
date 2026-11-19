package unknowninspection

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/pattern"
	"github.com/janalis/custos/internal/php/syntax"
)

// A project's custom rule is a known inspection name.
func TestCustomRuleKnown(t *testing.T) {
	custom, err := pattern.Compile(pattern.Definition{ID: "PhpNoDump", Match: pattern.Patterns{`\dump()`}, Message: "m"})
	if err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine([]analysis.Rule{New(), custom}, analysis.Config{Only: []string{"UnknownInspection"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("a.php", []byte("<?php\n/** @noinspection PhpNoDump PhpOther */\nf();"), syntax.Options{}))
	if len(got) != 1 || got[0].Message != "Suppressed inspection not recognized: PhpOther." {
		t.Fatalf("got %+v", got)
	}
}
