package alterinforeach

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
		{"empty statement", "<?php function f($a) { foreach ($a as &$v) {}; unset($v); }", 0},
		{"empty statement then use", "<?php function f($a) { foreach ($a as &$v) { g(); }; echo 1; }", 1},
		{"switch break", "<?php function f($a) { switch ($a) { case 1: foreach ($a as &$v) {} break; } unset($v); }", 0},
		{"loop break", "<?php function f($a) { while (1) { foreach ($a as &$v) {} break; } unset($v); }", 0},
		{"levelled break", "<?php function f($a) { while (1) { while (1) { foreach ($a as &$v) {} break 2; } } unset($v); }", 0},
		{"continue", "<?php function f($a) { while (1) { foreach ($a as &$v) {} continue; } unset($v); }", 0},
		{"outer loop writes", "<?php function f($a) { while (1) { $v = 0; foreach ($a as &$v) {} continue; } unset($v); }", 1},
		{"stray break", "<?php function f($a) { foreach ($a as &$v) { g(); } break; }", 1},
		{"earlier by-ref loop", "<?php function f($a, $b) { foreach ($a as &$v) {} foreach ($b as $v) {} unset($v); }", 1},
		{"by-ref parameter", "<?php function f($a, &$v) { foreach ($a as $v) {} unset($v); }", 0},
		{"by-value parameter", "<?php function f($a, $v) { foreach ($a as $v) {} unset($v); }", 1},
		{"global", "<?php function f($a) { global $v; foreach ($a as $v) {} unset($v); }", 0},
		{"static", "<?php function f($a) { static $v; foreach ($a as $v) {} unset($v); }", 0},
		{"reference assignment", "<?php function f($a) { $v = &$a[0]; foreach ($a as $v) {} unset($v); }", 0},
		{"reference source", "<?php function f($a) { $w = &$v; foreach ($a as $v) {} unset($v); }", 0},
		{"closure use", "<?php $f = function ($a) use (&$v) { foreach ($a as $v) {} unset($v); };", 0},
		{"nested closure use", "<?php function f($a) { $g = function () use (&$v) {}; foreach ($a as $v) {} unset($v); }", 0},
		{"nested closure parameter", "<?php function f($a) { $g = function (&$v) {}; foreach ($a as $v) {} unset($v); }", 1},
		{"binding after unset", "<?php function f($a) { foreach ($a as $v) {} unset($v); $v = &$a[0]; }", 1},
		{"binding after unset in loop", "<?php function f($a) { for (;;) { foreach ($a as $v) {} unset($v); $v = &$a[0]; } }", 0},
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
