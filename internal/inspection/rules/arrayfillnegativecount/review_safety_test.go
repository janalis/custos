package arrayfillnegativecount

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
		{"literal variable", "<?php function f() { $n = -1; return array_fill(0, $n, 0); }", 1},
		{"loop increment", "<?php function f($xs) { $n = -1; foreach ($xs as $x) { $n++; } return array_fill(0, $n, 0); }", 0},
		{"post increment", "<?php function f() { $n = -1; $n++; return array_fill(0, $n, 0); }", 0},
		{"pre increment", "<?php function f() { $n = -1; ++$n; return array_fill(0, $n, 0); }", 0},
		{"decrement", "<?php function f() { $n = 1; $n = -1; $n--; return array_fill(0, $n, 0); }", 0},
		{"incremented source", "<?php function f() { $a = -1; $a++; $n = $a; return array_fill(0, $n, 0); }", 0},
		{"cyclic copies", "<?php function f() { $n = -1; $m = $n; $m++; $n = $m; return array_fill(0, $n, 0); }", 0},
		{"unrelated increment", "<?php function f() { $n = -1; $i = 0; $i++; return array_fill(0, $n, 0); }", 1},
		{"element increment", "<?php function f($a) { $n = -1; $a[0]++; return array_fill(0, $n, 0); }", 1},
		{"reference capture", "<?php function f() { $n = -1; $g = function () use (&$n) { $n = 3; }; $g(); return array_fill(0, $n, 0); }", 0},
		{"value capture", "<?php function f() { $n = -1; $g = function () use ($n) { return $n; }; return array_fill(0, $n, $g); }", 1},
		{"referenced item", "<?php function f() { $n = -1; $box = [&$n]; $box[0] = 3; return array_fill(0, $n, 0); }", 0},
		{"extract", "<?php function f($a) { $n = -1; extract($a); return array_fill(0, $n, 0); }", 0},
		{"namespaced extract", "<?php namespace App; function extract($a) {} function f($a) { $n = -1; extract($a); return \\array_fill(0, $n, 0); }", 1},
		{"arrow increment", "<?php function f() { $n = -1; $g = fn() => $n++; return array_fill(0, $n, $g); }", 1},
		{"global declaration", "<?php $n = -1; g(); array_fill(0, $n, 0); function g() { global $n; $n = 3; }", 0},
		{"globals array", "<?php $n = -1; g(); array_fill(0, $n, 0); function g() { $GLOBALS['n'] = 3; }", 0},
		{"file scope literal", "<?php $n = -1; array_fill(0, $n, 0);", 1},
		{"literal count", "<?php function f() { $n = 0; $n++; return array_fill(0, -1, $n); }", 1},
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
