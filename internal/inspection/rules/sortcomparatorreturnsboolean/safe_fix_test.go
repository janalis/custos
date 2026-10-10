package sortcomparatorreturnsboolean

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestFixPreservesComments(t *testing.T) {
	for _, tc := range []struct{ source, fixed string }{
		{"usort($xs, fn($a,$b) => $a /* first */ > /* second */ $b);", "usort($xs, fn($a,$b) => $a /* first */ <=> /* second */ $b);"},
		{"usort($xs, fn($a,$b) => $a /* first */ < /* second */ $b);", "usort($xs, fn($a,$b) => $b /* first */ <=> /* second */ $a);"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SortComparatorReturnsBoolean"}})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte("<?php " + tc.source)
			findings := engine.Analyze(syntax.Parse("test.php", src, syntax.Options{}))
			if len(findings) != 1 || len(findings[0].Fixes) != 1 {
				t.Fatalf("expected one repair: %+v", findings)
			}
			fixed, _ := fixing.Apply(src, findings[0].Fixes[0].Edits())
			if string(fixed) != "<?php "+tc.fixed {
				t.Fatalf("fixed: %s", fixed)
			}
			file := syntax.Parse("fixed.php", fixed, syntax.Options{})
			if len(file.Errors) != 0 || len(engine.Analyze(file)) != 0 {
				t.Fatal("repair must parse and remove the finding")
			}
		})
	}
}
