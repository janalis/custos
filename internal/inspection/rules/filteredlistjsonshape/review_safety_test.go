package filteredlistjsonshape

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
	const use = " return json_encode(array_filter($r)); }"
	const item = " return json_encode(array_filter([1, $z, 2])); }"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"plain local", "<?php function f() { $r = [1, 0, 2];" + use, 1},
		{"top-level local", "<?php $r = [1, 0, 2]; json_encode(array_filter(($r)));", 1},
		{"by-value capture", "<?php function f() { $r = [1, 0, 2]; $l = function () use ($r) { $r = [3, 4]; }; $l();" + use, 1},
		{"capture of another variable", "<?php function f() { $r = [1, 0, 2]; $q = 1; $l = function () use (&$q) {}; $h = [&$q]; $q++;" + use, 1},
		{"write inside arrow function", "<?php function f() { $r = [1, 0, 2]; $g = fn() => $r++;" + use, 1},
		{"copied to array item and variable", "<?php function f() { $r = [1, 0, 2]; $h = [$r]; $c = $r; foreach ($r as $v) {}" + use, 1},
		{"extract in nested closure only", "<?php function f($o) { $r = [1, 0, 2]; $g = function () use ($o) { extract($o); };" + use, 0},
		{"variable item", "<?php function f() { $z = 0;" + item, 1},
		{"copied variable item", "<?php function f() { $y = 0; $z = $y;" + item, 1},
		{"variable key", "<?php function f() { $k = 1; return json_encode(array_filter([0 => 1, $k => 0, 2 => 2])); }", 1},
		{"by-reference capture", "<?php function f() { $r = [1, 0, 2]; $l = function () use (&$r) { $r = [3, 4]; }; $l();" + use, 0},
		{"referenced array item", "<?php function f() { $r = [1, 0, 2]; $h = [&$r]; $h[0] = [3, 4];" + use, 0},
		{"reference assignment source", "<?php function f() { $r = [1, 0, 2]; $q = &$r; $q = [3, 4];" + use, 0},
		{"reference assignment target", "<?php function f($o) { $r = [1, 0, 2]; $r = &$o;" + use, 0},
		{"foreach by reference", "<?php function f() { $r = [1, 0, 2]; foreach ($r as &$v) { $v = 1; }" + use, 0},
		{"foreach target", "<?php function f($o) { $r = [1, 0, 2]; foreach ($o as $r) {}" + use, 0},
		{"catch variable", "<?php function f() { $r = [1, 0, 2]; try { g(); } catch (E $r) {}" + use, 0},
		{"extract", "<?php function f($o) { $r = [1, 0, 2]; extract($o);" + use, 0},
		{"top-level extract", "<?php $r = [1, 0, 2]; \\extract($o); json_encode(array_filter($r));", 0},
		{"incremented item", "<?php function f() { $z = 0; $z++;" + item, 0},
		{"captured item", "<?php function f() { $z = 0; $l = function () use (&$z) { $z = 5; }; $l();" + item, 0},
		{"captured copy source", "<?php function f() { $y = 0; $l = function () use (&$y) { $y = 4; }; $l(); $z = $y;" + item, 0},
		{"incremented key", "<?php function f() { $k = 1; $k++; return json_encode(array_filter([0 => 1, $k => 0, 2 => 2])); }", 0},
		{"item incremented after encoding", "<?php function f() { $z = 0; $j = json_encode(array_filter([1, $z, 2])); $z++; return $j; }", 0},
		{"repeated variable item", "<?php function f() { $z = 0; return json_encode(array_filter([$z, 1, $z])); }", 1},
		{"mutually assigned locals", "<?php function f($n) { $a = [1, 0, 2]; $b = $a; if ($n) { $a = $b; } return json_encode(array_filter($b)); }", 1},
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
