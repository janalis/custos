package gdpngcompressionoutsidesupportedrange

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
		{"plain literal", "<?php function f($i, $p) { imagepng($i, $p, 10); }", 1},
		{"plain local", "<?php function f($i, $p) { $l = 10; imagepng($i, $p, $l); }", 1},
		{"top-level local", "<?php $l = 10; imagepng($i, $p, $l);", 1},
		{"named quality", "<?php function f($i, $p) { $l = 10; imagepng(quality: $l, image: $i, file: $p); }", 1},
		{"below sentinel", "<?php function f($i, $p) { $l = -2; imagepng($i, $p, $l); }", 1},
		{"by-value capture", "<?php function f($i, $p) { $l = 10; $g = function () use ($l) {}; imagepng($i, $p, $l); }", 1},
		{"other variable decremented", "<?php function f($i, $p) { $l = 10; $t = 10; $t--; imagepng($i, $p, $l); }", 1},
		{"decrement in nested closure", "<?php function f($i, $p) { $l = 10; $g = function () { $l = 1; $l--; }; imagepng($i, $p, $l); }", 1},
		{"array item by value", "<?php function f($i, $p) { $l = 10; $b = [$l]; imagepng($i, $p, $l); }", 1},
		{"post decrement", "<?php function f($i, $p) { $l = 10; $l--; imagepng($i, $p, $l); }", 0},
		{"pre decrement", "<?php function f($i, $p) { $l = 10; --$l; imagepng($i, $p, $l); }", 0},
		{"increment from below", "<?php function f($i, $p) { $l = -2; $l++; imagepng($i, $p, $l); }", 0},
		{"loop decrement", "<?php function f($i, $p, $xs) { $l = 10; foreach ($xs as $x) { $l--; } imagepng($i, $p, $l); }", 0},
		{"decrement after call in loop", "<?php function f($i, $p) { $l = 10; while (g()) { $l--; imagepng($i, $p, $l); } }", 0},
		{"top-level decrement", "<?php $l = 10; $l--; imagepng($i, $p, $l);", 0},
		{"arithmetic on decremented", "<?php function f($i, $p) { $l = 9; $l--; imagepng($i, $p, $l + 1); }", 0},
		{"compound assignment", "<?php function f($i, $p) { $l = 10; $l -= 1; imagepng($i, $p, $l); }", 0},
		{"by-reference capture", "<?php function f($i, $p) { $l = 10; $g = function () use (&$l) { $l = 3; }; $g(); imagepng($i, $p, $l); }", 0},
		{"referenced array item", "<?php function f($i, $p) { $l = 10; foreach ([&$l] as &$q) { $q = 3; } imagepng($i, $p, $l); }", 0},
		{"extract", "<?php function f($i, $p, $o) { $l = 10; extract($o); imagepng($i, $p, $l); }", 0},
		{"top-level extract", "<?php $l = 10; \\extract($o); imagepng($i, $p, $l);", 0},
		{"unresolved callee", "<?php function f($i, $p) { $l = 10; adjust($l); imagepng($i, $p, $l); }", 0},
		{"by-reference parameter", "<?php function adjust(&$v) { $v = 3; } function f($i, $p) { $l = 10; adjust($l); imagepng($i, $p, $l); }", 0},
		{"method argument", "<?php function f($i, $p, $o) { $l = 10; $o->adjust($l); imagepng($i, $p, $l); }", 0},
		{"static method argument", "<?php function f($i, $p) { $l = 10; Levels::adjust($l); imagepng($i, $p, $l); }", 0},
		{"constructor argument", "<?php function f($i, $p) { $l = 10; new Adjuster($l); imagepng($i, $p, $l); }", 0},
		{"nullsafe method argument", "<?php function f($i, $p, $o) { $l = 10; $o?->adjust($l); imagepng($i, $p, $l); }", 0},
		{"dynamic callee", "<?php function f($i, $p, $g) { $l = 10; $g($l); imagepng($i, $p, $l); }", 0},
		{"by-value parameter", "<?php function show($v) {} function f($i, $p) { $l = 10; show($l); imagepng($i, $p, $l); }", 1},
		{"builtin by-value argument", "<?php function f($i, $p) { $l = 10; var_dump($l); imagepng($i, $p, $l); }", 1},
		{"repeated call", "<?php function f($i, $p) { $l = 10; imagepng($i, $p, $l); imagepng($i, $p, $l); }", 2},
		{"arrow function scope", "<?php function f($i, $p) { $l = 10; $l--; return fn() => imagepng($i, $p, $l); }", 0},
		{"reference alias", "<?php function f($i, $p) { $l = 10; $r = &$l; $r = 3; imagepng($i, $p, $l); }", 0},
		{"branch assignment", "<?php function f($i, $p, $c) { $l = 10; if ($c) { $l = 3; } imagepng($i, $p, $l); }", 0},
		{"foreach value", "<?php function f($i, $p, $xs) { $l = 10; foreach ($xs as $l) {} imagepng($i, $p, $l); }", 0},
		{"destructuring", "<?php function f($i, $p, $xs) { $l = 10; [$l] = $xs; imagepng($i, $p, $l); }", 0},
		{"variable variable", "<?php function f($i, $p) { $l = 10; $n = 'l'; $$n = 3; imagepng($i, $p, $l); }", 0},
		{"output argument", "<?php function f($i, $p) { $l = 10; preg_match('/x/', 'x', $l); imagepng($i, $p, $l); }", 0},
		{"unset", "<?php function f($i, $p) { $l = 10; unset($l); imagepng($i, $p, $l); }", 0},
		{"static", "<?php function f($i, $p) { static $l = 10; imagepng($i, $p, $l); $l = 3; }", 0},
		{"global", "<?php function f($i, $p) { $l = 10; global $l; imagepng($i, $p, $l); }", 0},
		{"copied after decrement", "<?php function f($i, $p) { $b = 10; $b--; $l = $b; imagepng($i, $p, $l); }", 0},
		{"copied twice after capture", "<?php function f($i, $p) { $b = 10; $g = function () use (&$b) { $b = 3; }; $g(); $m = $b; $l = $m; imagepng($i, $p, $l); }", 0},
		{"copied plain", "<?php function f($i, $p) { $b = 10; $l = $b; $b = $l; imagepng($i, $p, $l); }", 1},
		{"arrow function decrement", "<?php function f($i, $p) { $l = 10; $g = fn() => $l--; imagepng($i, $p, $l); }", 1},
		{"decrement after call", "<?php function f($i, $p) { $l = 10; imagepng($i, $p, $l); $l--; }", 0},
		{"by-reference capture never invoked", "<?php function f($i, $p) { $l = 10; $g = function () use (&$l) { $l = 3; }; imagepng($i, $p, $l); }", 0},
		{"referenced array item never written", "<?php function f($i, $p) { $l = 10; $b = [&$l]; imagepng($i, $p, $l); }", 0},
		{"extract after call", "<?php function f($i, $p, $o) { $l = 10; imagepng($i, $p, $l); extract($o); }", 0},
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
