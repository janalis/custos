package explodeemptyseparator

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
		{"local literal", "<?php function f($s) { $sep = ''; return explode($sep, $s); }", 1},
		{"file literal", "<?php $sep = ''; explode($sep, $s);", 1},
		{"file literal beside a function", "<?php function g() { $x = 1; $x++; } $sep = ''; explode($sep, $s);", 1},
		{"by-reference capture", "<?php function f($s) { $sep = ''; $c = function () use (&$sep) { $sep = ','; }; $c(); return explode($sep, $s); }", 0},
		{"by-value capture", "<?php function f($s) { $sep = ''; $c = function () use ($sep) { $sep = ','; }; $c(); return explode($sep, $s); }", 1},
		{"capturing scope", "<?php $c = function ($s) use (&$sep) { $sep = ''; refill(); return explode($sep, $s); };", 0},
		{"global in function", "<?php function f($s) { global $sep; $sep = ''; configure(); return explode($sep, $s); }", 0},
		{"other global in function", "<?php function f($s) { global $other, $$name; $sep = ''; return explode($sep, $s); }", 1},
		{"file global written elsewhere", "<?php $sep = ''; function configure() { global $sep; $sep = ','; } configure(); explode($sep, $s);", 0},
		{"file global through superglobal", "<?php $sep = ''; function configure() { $GLOBALS['sep'] = ','; } configure(); explode($sep, $s);", 0},
		{"extract", "<?php function f($s, $o) { $sep = ''; extract($o); return explode($sep, $s); }", 0},
		{"namespaced extract", "<?php namespace App; function extract($o) {} function f($s, $o) { $sep = ''; extract($o); return \\explode($sep, $s); }", 1},
		{"parse_str into locals", "<?php function f($s, $q) { $sep = ''; parse_str($q); return explode($sep, $s); }", 0},
		{"parse_str into result", "<?php function f($s, $q) { $sep = ''; parse_str($q, $r); return explode($sep, $s); }", 1},
		{"increment", "<?php function f($s) { $sep = ''; $sep++; return explode($sep, $s); }", 0},
		{"property increment", "<?php function f($s, $o) { $sep = ''; $o->n++; return explode($sep, $s); }", 1},
		{"referenced array item", "<?php function f($s) { $sep = ''; $a = array('k' => &$sep); fill($a); return explode($sep, $s); }", 0},
		{"copied from captured", "<?php function f($s) { $a = ''; $c = function () use (&$a) { $a = ','; }; $c(); $sep = $a; return explode($sep, $s); }", 0},
		{"copy cycle", "<?php function f($s) { $a = ''; $sep = $a; $sep = $sep; return explode($sep, $s); }", 1},
		{"named argument", "<?php function f($s) { $sep = ''; $sep++; return explode(string: $s, separator: $sep); }", 0},
		{"literal argument", "<?php function f($s, $o) { extract($o); return explode('', $s); }", 1},
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
