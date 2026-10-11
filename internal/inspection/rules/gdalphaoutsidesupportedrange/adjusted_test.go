package gdalphaoutsidesupportedrange

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAdjustedBefore(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"file scope literal", "$a = 200; imagecolorallocatealpha($i, 1, 2, 3, $a);", 1},
		{"copied literal", "function f($i) { $b = 200; $a = $b; return imagecolorallocatealpha($i, 1, 2, 3, $a); }", 1},
		{"builtin call before", "function f($i) { $a = 200; strlen('x'); return imagecolorallocatealpha($i, 1, 2, 3, $a); }", 1},
		{"reassigned after step", "function f($i) { $a = 1; $a++; $a = 200; return imagecolorallocatealpha($i, 1, 2, 3, $a); }", 0},
		{"reference capture never invoked", "function f($i) { $a = 200; $g = function () use (&$a) { $a = 3; }; return imagecolorallocatealpha($i, 1, 2, $g, $a); }", 0},
		{"reassigned after extract", "function f($i, $o) { extract($o); $a = 200; return imagecolorallocatealpha($i, 1, 2, 3, $a); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GdAlphaOutsideSupportedRange"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("adjusted.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
