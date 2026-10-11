package arrayuniquecanbeused

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
		{"array_push", "<?php function f() { $a = [1, 2]; array_push($a, '7', '2'); return array_keys(array_count_values($a)); }", 0},
		{"foreach by reference", "<?php function f() { $a = [1, 2]; foreach ($a as &$v) { $v = \"$v\"; } return array_keys(array_count_values($a)); }", 0},
		{"keyed foreach by reference", "<?php function f() { $a = [1, 2]; foreach ($a as $k => &$v) { $v /= 4; } return count(array_count_values($a)); }", 0},
		{"alias", "<?php function f() { $a = [1, 2]; $r = &$a; $r[] = 1.5; return count(array_count_values($a)); }", 0},
		{"alias kept", "<?php function f() { $a = [1, 2]; $r = &$a; $b = 1; $a = [1, 2]; g($r); return count(array_count_values($a)); }", 0},
		{"element alias", "<?php function f() { $a = [1, 2]; $r = &$a[0]; $a = [1, 2]; $r = 1.5; return count(array_count_values($a)); }", 0},
		{"closure by reference", "<?php function f() { $a = [1, 2]; $g = function () use (&$a) { $a = [1.5, 1.7]; }; $g(); return count(array_count_values($a)); }", 0},
		{"closure late", "<?php function f() { $g = function () use (&$a) { $a = [1.5, 1.7]; }; $a = [1, 2]; $g(); return count(array_count_values($a)); }", 0},
		{"array_walk", "<?php function f() { $a = [1, 2]; array_walk($a, function (&$v) { $v /= 4; }); return count(array_count_values($a)); }", 0},
		{"unknown callee", "<?php function f() { $a = [1, 2]; g($a); return count(array_count_values($a)); }", 0},
		{"method callee", "<?php function f($o) { $a = [1, 2]; $o->g($a); return count(array_count_values($a)); }", 0},
		{"element argument", "<?php function f() { $a = [1, 2]; settype($a[0], 'float'); return count(array_count_values($a)); }", 0},
		{"referenced item", "<?php function f() { $a = [1, 2]; $b = [&$a]; $b[0][] = 1.5; return count(array_count_values($a)); }", 0},
		{"foreach target", "<?php function f($x) { $a = [1, 2]; foreach ($x as $a[0]) {} return count(array_count_values($a)); }", 0},
		{"global", "<?php function f() { $a = [1, 2]; global $a; return count(array_count_values($a)); }", 0},
		{"static", "<?php function f() { $a = [1, 2]; static $a; return count(array_count_values($a)); }", 0},
		{"variable variable", "<?php function f($n) { $a = [1, 2]; $$n = [1.5]; return count(array_count_values($a)); }", 0},
		{"extract", "<?php function f($n) { $a = [1, 2]; extract($n); return count(array_count_values($a)); }", 0},
		{"unset element", "<?php function f() { $a = [1, 2]; unset($a[0]); return count(array_count_values($a)); }", 1},
		{"by-value builtin", "<?php function f() { $a = [1, 2]; echo count($a); return count(array_count_values($a)); }", 1},
		{"by-value foreach", "<?php function f() { $a = [1, 2]; foreach ($a as $v) { echo $v; } return count(array_count_values($a)); }", 1},
		{"variadic", "<?php function f(int ...$n) { array_push($n, 1.5); return count(array_count_values($n)); }", 0},
		{"variadic safe", "<?php function f(int ...$n) { return array_keys(array_count_values($n)); }", 1},
		{"reference parameter", "<?php function f(&$a) { $a = [1, 2]; g(); return count(array_count_values($a)); }", 0},
		{"static callee", "<?php function f() { $a = [1, 2]; K::g(($a)); return count(array_count_values($a)); }", 0},
		{"constructor", "<?php function f() { $a = [1, 2]; new K($a); return count(array_count_values($a)); }", 0},
		{"include", "<?php function f() { $a = [1, 2]; include 'x.php'; return count(array_count_values($a)); }", 0},
		{"eval", "<?php function f($c) { $a = [1, 2]; eval($c); return count(array_count_values($a)); }", 0},
		{"unrelated calls", "<?php function f($b) { $a = [1, 2]; g($b); $h = strlen(...); new K; return count(array_count_values($a)); }", 1},
		{"top level", "<?php $a = [1, 2]; array_push($a, 1.5); echo count(array_count_values($a));", 0},
		{"top level safe", "<?php $a = [1, 2]; echo count(array_count_values($a));", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
