package sortcomparatorreturnsboolean

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
		name, source string
		findings     int
		fixed        string
	}{
		{"inline closure", "<?php usort($xs, function ($a, $b) { return $a > $b; });", 1, "<?php usort($xs, function ($a, $b) { return $a <=> $b; });"},
		{"inline int arrow", "<?php usort($xs, fn(int $a, int $b): int => $a < $b);", 1, "<?php usort($xs, fn(int $a, int $b): int => $b <=> $a);"},
		{"parenthesized literal", "<?php usort($xs, (fn($a, $b) => $a > $b));", 1, "<?php usort($xs, (fn($a, $b) => $a <=> $b));"},
		{"named callback argument", "<?php usort(callback: fn($a, $b) => $a > $b, array: $xs);", 1, "<?php usort(callback: fn($a, $b) => $a <=> $b, array: $xs);"},
		{"bool closure", "<?php usort($xs, function (int $a, int $b): bool { return $a > $b; });", 1, ""},
		{"bool arrow", "<?php usort($xs, fn(int $a, int $b): bool => $a < $b);", 1, ""},
		{"nullable int", "<?php usort($xs, fn($a, $b): ?int => $a > $b);", 1, ""},
		{"union type", "<?php usort($xs, fn($a, $b): int|bool => $a > $b);", 1, ""},
		{"named function", "<?php function isNewer($a, $b) { return $a > $b; } usort($xs, 'isNewer'); isNewer(1, 2);", 1, ""},
		{"closure variable", "<?php $gt = fn($a, $b) => $a > $b; usort($xs, $gt); $gt(1, 2);", 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
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
