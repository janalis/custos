package callbackrequiredargumentmismatch

import (
	"testing"

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
	}{
		{"unique declaration", "<?php function f($a, $b) {} array_map('f', $xs);", 1},
		{"declared after use", "<?php array_map('f', $xs); function f($a, $b) {}", 1},
		{"namespace statement", "<?php namespace App; function f($a, $b) {} array_map('App\\\\f', $xs);", 1},
		{"braced namespace", "<?php namespace App { function f($a, $b) {} array_map('App\\\\f', $xs); }", 1},
		{"unique closure in branch", "<?php if ($c) { array_map(function ($a, $b) {}, $xs); }", 1},
		{"branch declarations", "<?php if (PHP_VERSION_ID < 80000) { function f($a, $b) {} } else { function f($a) {} } array_map('f', [1]);", 0},
		{"reversed branches", "<?php if (PHP_VERSION_ID < 80000) { function f($a) {} } else { function f($a, $b) {} } array_map('f', [1]);", 0},
		{"guarded fallback", "<?php if (!function_exists('f')) { function f($a, $b) {} } array_map('f', $xs);", 0},
		{"nested in function", "<?php function outer() { function f($a, $b) {} } array_map('f', $xs);", 0},
		{"mixed case duplicate", "<?php function f($a, $b) {} if ($c) { function F($a) {} } array_map('f', $xs);", 0},
		{"enough arrays", "<?php function f($a, $b) {} array_map('f', $xs, $ys);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
