package usinginclusiononcereturnvalue

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
		{"global", "<?php function f($f) { global $c; $c = include_once $f; }", 1},
		{"static", "<?php function f($f) { static $c; $c = include_once $f; }", 1},
		{"static tested", "<?php function f($f) { static $c; $c = include_once $f; if (!$c) { die(); } }", 1},
		{"by-ref param", "<?php function f($f, &$out) { $out = include_once $f; }", 1},
		{"by-ref method param", "<?php class A { function f($f, &$out) { $out = require_once $f; } }", 1},
		{"by-value param", "<?php function f($f, $out) { $out = include_once $f; }", 0},
		{"by-ref closure use", "<?php $g = function ($f) use (&$c) { $c = include_once $f; };", 1},
		{"by-value closure use", "<?php $g = function ($f) use ($c) { $c = include_once $f; };", 0},
		{"compact", "<?php function f($f) { $x = include_once $f; return compact('x'); }", 1},
		{"compact case", "<?php function f($f) { $x = include_once $f; return \\COMPACT('x'); }", 1},
		{"get_defined_vars", "<?php function f($f) { $x = include_once $f; return get_defined_vars(); }", 1},
		{"variable-variable", "<?php function f($f) { $x = include_once $f; $n = 'x'; return $$n; }", 1},
		{"eval", "<?php function f($f, $c) { $x = include_once $f; if ($x) { eval($c); } }", 1},
		{"template", "<?php function f($f) { $x = include_once $f; include 'tpl.php'; }", 1},
		{"template in arrow", "<?php function f($f) { $x = include_once $f; return fn() => include 'tpl.php'; }", 1},
		{"tested with fallback", "<?php function f($a, $b) { $r = @include_once $a; if (!$r) { $r = @include_once $b; } }", 0},
		{"nested scope by name", "<?php function f($f) { $x = include_once $f; return function ($n) { return $$n; }; }", 0},
		{"nested function", "<?php function f($f) { $x = include_once $f; function g() { return get_defined_vars(); } }", 0},
		{"nested class", "<?php function f($f) { $x = include_once $f; return new class { function g() { include 'a.php'; } }; }", 0},
		{"file scope by name", "<?php $x = include_once 'a.php'; if (!$x) { exit; } echo $$n;", 1},
		{"file scope flag", "<?php $x = include_once 'a.php'; if (!$x) { exit; } include 'b.php';", 0},
		{"unread local", "<?php function f($f) { $unused = include_once $f; }", 0},
		{"dynamic call", "<?php function f($f, $g) { $x = include_once $f; $g(); }", 0},
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
