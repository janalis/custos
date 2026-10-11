package sortcomparatorreturnsfloat

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		php          phpversion.Version
		findings     int
		fixed        string
	}{
		{"inline arrow", "<?php usort($xs, fn(float $a, float $b) => $a - $b);", phpversion.PHP85, 1, "<?php usort($xs, fn(float $a, float $b) => $a <=> $b);"},
		{"inline closure", "<?php usort($xs, function (float $a, float $b) { return $a - $b; });", phpversion.PHP85, 1, "<?php usort($xs, function (float $a, float $b) { return $a <=> $b; });"},
		{"parenthesized closure", "<?php usort($xs, (function (float $a, float $b) { return $a - $b; }));", phpversion.PHP85, 1, "<?php usort($xs, (function (float $a, float $b) { return $a <=> $b; }));"},
		{"named function", "<?php function diff(float $a, float $b) { return $a - $b; } usort($xs, 'diff'); echo diff(2.5, 1.0);", phpversion.PHP85, 1, ""},
		{"closure variable", "<?php $delta = fn(float $a, float $b) => $a - $b; usort($xs, $delta); echo $delta(2.5, 1.0);", phpversion.PHP85, 1, ""},
		{"two returns", "<?php usort($xs, function (float $a, float $b) { if ($a > 1.0) { return $b - $a; } return $a - $b; });", phpversion.PHP85, 1, ""},
		{"no spaceship", "<?php usort($xs, function ($a, $b) { return (float) $a - (float) $b; });", phpversion.PHP56, 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte(tc.source)
			opt := syntax.Options{Version: tc.php}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
