package fractionalmodulooperand

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
		{"local literal", "<?php function f($a) { $s = 1.5; return $a % $s; }", 1},
		{"reference capture", "<?php function f($a) { $s = 1.5; $g = function () use (&$s) { $s = 3; }; $g(); return $a % $s; }", 0},
		{"chained reference capture", "<?php function f($a) { $b = 1.5; $g = function () use (&$b) { $b = 3; }; $g(); $s = $b; return $a % $s; }", 0},
		{"value capture", "<?php function f($a) { $s = 1.5; $g = function () use ($a, $s) { $s = 3; }; $g(); return $a % $s; }", 1},
		{"capture after use", "<?php function f($a) { $s = 1.5; $r = $a % $s; $g = function () use (&$s) { $s = 3; }; return $r; }", 1},
		{"capture after use in loop", "<?php function f($a) { $s = 1.5; for (;;) { $r = $a % $s; $g = function () use (&$s) { $s = 3; }; $g(); } }", 0},
		{"nested scope capture", "<?php function f($a) { $s = 1.5; $g = function () { $s = 1; return function () use (&$s) {}; }; return $a % $s; }", 1},
		{"extract", "<?php function f($a, $o) { $s = 0.5; extract($o); return $a % $s; }", 0},
		{"qualified extract", "<?php namespace App; function f($a, $o) { $s = 0.5; \\extract($o); return $a % $s; }", 0},
		{"extract after use", "<?php function f($a, $o) { $s = 0.5; $r = $a % $s; extract($o); return $r; }", 1},
		{"extract after use in foreach", "<?php function f($a, $o) { $s = 0.5; foreach ($o as $x) { $r = $a % $s; extract($o); } }", 0},
		{"repeated use, nothing written", "<?php function f($a) { $s = 0.5; while ($a) { $r = $a % $s; $a--; } }", 1},
		{"extract after use in do", "<?php function f($a, $o) { $s = 0.5; do { $r = $a % $s; extract($o); } while ($a); }", 0},
		{"user extract", "<?php namespace App; function extract($o) {} function f($a, $o) { $s = 0.5; extract($o); return $a % $s; }", 1},
		{"parse_str into scope", "<?php function f($a, $q) { $s = 0.5; parse_str($q); return $a % $s; }", 0},
		{"parse_str into result", "<?php function f($a, $q) { $s = 0.5; parse_str($q, $out); return $a % $s; }", 1},
		{"goto", "<?php function f($a) { $s = 0.5; $r = $a % $s; again: $s = 2; goto again; }", 0},
		{"local user call", "<?php function f($a) { $s = 0.5; g(); $o = new K(); $o->m(); K::n(); return $a % $s; }", 1},
		{"literal operand beside extract", "<?php function f($a, $o) { extract($o); return $a % 0.5; }", 1},
		{"global literal", "<?php $s = 0.5; echo strlen('a'); echo 9 % $s;", 1},
		{"global user call", "<?php $s = 0.5; g(); echo 9 % $s;", 0},
		{"global construction", "<?php $s = 0.5; $o = new K(); echo 9 % $s;", 0},
		{"global method call", "<?php $s = 0.5; $o->m(); echo 9 % $s;", 0},
		{"global static call", "<?php $s = 0.5; K::m(); echo 9 % $s;", 0},
		{"global call after use", "<?php $s = 0.5; echo 9 % $s; g(); $o = new K(); $o->m(); K::m();", 1},
		{"global call in declaration", "<?php function g() { h(); } $s = 0.5; echo 9 % $s;", 1},
		{"arrow scope", "<?php $f = fn($a) => (function () use (&$a) {})() || $a % 0.5;", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("test.php", []byte(tc.source), syntax.Options{}))
			if len(got) != tc.findings {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.findings, got)
			}
		})
	}
}
