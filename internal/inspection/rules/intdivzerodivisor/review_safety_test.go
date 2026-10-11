package intdivzerodivisor

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
		{"known zero", "<?php function f() { $d = 0; return intdiv(6, $d); }", 1},
		{"post increment", "<?php function f() { $d = 0; $d++; return intdiv(6, $d); }", 0},
		{"pre increment", "<?php function f() { $d = 0; ++$d; return intdiv(6, $d); }", 0},
		{"decrement", "<?php function f() { $d = 0; $d--; return intdiv(6, $d); }", 0},
		{"loop counter", "<?php function f($xs) { $d = 0; foreach ($xs as $x) { $d++; echo intdiv($x, $d); } }", 0},
		{"for counter", "<?php function f() { for ($d = 0; $d < 3; $d++) { if ($d) { echo intdiv(6, $d); } } }", 0},
		{"stepped later in loop", "<?php function f() { $d = 0; while (c()) { if ($d) { echo intdiv(6, $d); } $d++; } }", 0},
		{"do while", "<?php function f() { $d = 0; do { $d++; } while (c()); return intdiv(6, $d); }", 0},
		{"reference capture", "<?php function f() { $d = 0; $s = function () use (&$d) { $d = 2; }; $s(); return intdiv(6, $d); }", 0},
		{"value capture", "<?php function f() { $d = 0; $s = function () use ($d) { return $d; }; return intdiv(6, $d); }", 1},
		{"stepped afterwards", "<?php function f() { $d = 0; $r = intdiv(6, $d); $d++; return $r; }", 1},
		{"other counter", "<?php function f() { $d = 0; $i = 0; $i++; return intdiv(6, $d); }", 1},
		{"property counter", "<?php function f($o) { $d = 0; $o->d++; return intdiv(6, $d); }", 1},
		{"other scope", "<?php function g() { $d = 0; $d++; } function f() { $d = 0; return intdiv(6, $d); }", 1},
		{"destructuring foreach", "<?php function f($xs) { $d = 0; foreach ($xs as [$n, $d]) { echo intdiv($n, $d); } }", 0},
		{"catch binding", "<?php function f() { $d = 0; try { g(); } catch (E $d) { } return intdiv(6, $d); }", 0},
		{"extract", "<?php function f($xs) { $d = 0; extract($xs); return intdiv(6, $d); }", 0},
		{"nested closure counter", "<?php function f() { $d = 0; $g = function () { $d = 1; $d++; extract([]); }; return intdiv(6, $d); }", 1},
		{"nested reference capture", "<?php function f() { $d = 0; $g = function () { $d = 1; $h = function () use (&$d) { }; }; return intdiv(6, $d); }", 1},
		{"top level beside function counter", "<?php function g() { $d = 1; $d++; } $d = 0; intdiv(6, $d);", 1},
		{"top level", "<?php $d = 0; $d++; intdiv(6, $d);", 0},
		{"arrow function", "<?php $d = 0; $d++; $f = fn () => intdiv(6, 0);", 1},
		{"method", "<?php class C { function f() { $d = 0; $d++; return intdiv(6, $d); } }", 0},
		{"unchanged in loop", "<?php function f($xs) { $d = 0; foreach ($xs as $x) { echo intdiv($x, $d); } }", 1},
		{"reference capture before any call", "<?php function f() { $d = 0; $s = function () use (&$d) { $d = 2; }; return intdiv(6, $d); }", 0},
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
