package arraychunkinvalidsize

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
		{"plain zero", "<?php function f($a) { $s = 0; return array_chunk($a, $s); }", 1},
		{"top-level zero", "<?php $s = 0; array_chunk($a, $s);", 1},
		{"named size", "<?php function f($a) { $s = 0; return array_chunk(length: $s, array: $a); }", 1},
		{"by-value capture", "<?php function f($a) { $s = 0; $g = function () use ($s) {}; return array_chunk($a, $s); }", 1},
		{"other variable incremented", "<?php function f($a) { $s = 0; $t = 0; $t++; return array_chunk($a, $s); }", 1},
		{"increment in nested closure", "<?php function f($a) { $s = 0; $g = function () { $s = 1; $s++; }; return array_chunk($a, $s); }", 1},
		{"post increment", "<?php function f($a) { $s = 0; $s++; return array_chunk($a, $s); }", 0},
		{"pre increment", "<?php function f($a) { $s = 0; ++$s; return array_chunk($a, $s); }", 0},
		{"loop increment", "<?php function f($a, $xs) { $s = 0; foreach ($xs as $x) { $s++; } return array_chunk($a, $s); }", 0},
		{"decrement", "<?php function f($a) { $s = 0; $s--; return array_chunk($a, $s); }", 0},
		{"top-level increment", "<?php $s = 0; $s++; array_chunk($a, $s);", 0},
		{"by-reference capture", "<?php function f($a) { $s = 0; $g = function () use (&$s) { $s = 3; }; $g(); return array_chunk($a, $s); }", 0},
		{"referenced array item", "<?php function f($a) { $s = 0; foreach ([&$s] as &$q) { $q = 3; } return array_chunk($a, $s); }", 0},
		{"array item by value", "<?php function f($a) { $s = 0; $b = [$s]; return array_chunk($a, $s); }", 1},
		{"extract", "<?php function f($a, $o) { $s = 0; extract($o); return array_chunk($a, $s); }", 0},
		{"top-level extract", "<?php $s = 0; \\extract($o); array_chunk($a, $s);", 0},
		{"copied after increment", "<?php function f($a, $xs) { $n = 0; foreach ($xs as $x) { $n++; } $s = $n; return array_chunk($a, $s); }", 0},
		{"copy chain after increment", "<?php function f($a) { $n = 0; $n++; $t = $n; $u = $t; return array_chunk($a, $u); }", 0},
		{"cyclic copies", "<?php function f($a) { $s = 0; $t = $s; $s = $t; return array_chunk($a, $s); }", 1},
		{"copied literal", "<?php function f($a) { $n = 0; $s = $n; return array_chunk($a, $s); }", 1},
		{"compound assignment", "<?php function f($a) { $s = 0; $s += 3; return array_chunk($a, $s); }", 0},
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
