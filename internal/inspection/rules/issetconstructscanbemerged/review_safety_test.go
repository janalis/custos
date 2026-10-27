package issetconstructscanbemerged

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, fixed string
	}{
		{"one line", "<?php if (isset($a) && isset($b) && $c) {}", "<?php if (isset($a, $b) && $c) {}"},
		{
			"trailing operators", "<?php\nif (isset($a) &&\n    isset($b) &&\n    ($c || $d)) {}",
			"<?php\nif (isset($a, $b) &&\n    ($c || $d)) {}",
		},
		{
			"leading operators", "<?php\nif (!isset($a)\n    || !isset($b)\n    || $c) {}",
			"<?php\nif (!isset($a, $b)\n    || $c) {}",
		},
		{
			"tab indentation", "<?php\nif (isset($a)\n\t&& $c\n\t&& isset($b)) {}",
			"<?php\nif (isset($a, $b)\n\t&& $c) {}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if string(res.Source) != tc.fixed {
				t.Fatalf("fixed source: got %q, want %q", res.Source, tc.fixed)
			}
		})
	}
}
