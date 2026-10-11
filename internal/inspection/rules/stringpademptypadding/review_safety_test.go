package stringpademptypadding

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewStaleArguments(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"length decremented", "<?php function f() { $w = 3; $w--; $w--; return str_pad('ab', $w, ''); }", 0},
		{"length incremented, parenthesized", "<?php function f() { $w = 9; ++$w; return str_pad('ab', ($w), ''); }", 0},
		{"pad written through reference capture", "<?php function f() { $p = ''; $g = function () use (&$p) { $p = '*'; }; $g(); return str_pad('ab', 5, $p); }", 0},
		{"input stepped", "<?php function f() { $s = 'a'; $s++; return str_pad($s, 5, ''); }", 0},
		{"copy of a stepped variable", "<?php function f() { $w = 3; $w--; $w--; $x = $w; return str_pad('ab', $x, ''); }", 0},
		{"stepped inside an arrow function", "<?php function f() { return fn() => [$w = 3, $w--, $w--, str_pad('ab', $w, '')]; }", 0},
		{"assigned inside an arrow function", "<?php function f() { return fn() => [$w = 9, str_pad('ab', $w, '')]; }", 1},
		{"step in the same loop", "<?php function f() { $w = 9; while (rand()) { echo str_pad('ab', $w, ''); $w--; } }", 0},
		{"step in a nested loop of the same loop", "<?php function f() { $w = 9; foreach ([1] as $_) { if (rand()) { echo str_pad('ab', $w, ''); } for (;;) { $w--; } } }", 0},
		{"extract before", "<?php function f() { $p = ''; extract(['p' => '*']); return str_pad('ab', 5, $p); }", 0},
		{"top-level extract before", "<?php $p = ''; extract(['p' => '*']); echo str_pad('ab', 5, $p);", 0},
		{"top-level decrement", "<?php $w = 3; $w--; $w--; echo str_pad('ab', $w, '');", 0},
		{"reference capture before any call", "<?php function f() { $p = ''; $g = function () use (&$p) { $p = '*'; }; return str_pad('ab', 5, $p); }", 0},
		{"step in another branch", "<?php function f($c) { if ($c) { $w = 3; $w--; return $w; } $w = 9; return str_pad('ab', $w, ''); }", 0},
		{"extract in another branch", "<?php function f($c, $xs) { if ($c) { extract($xs); return $p; } $p = ''; return str_pad('ab', 5, $p); }", 0},
		{"copy chain too long", "<?php function f() { $a = 9; $b = $a; $c = $b; $d = $c; $e = $d; $f = $e; $g = $f; $h = $g; $i = $h; $j = $i; return str_pad('ab', $j, ''); }", 0},

		{"plain variables", "<?php function f() { $w = 9; $p = ''; $s = 'ab'; return str_pad($s, $w, $p); }", 1},
		{"short copy chain", "<?php function f() { $a = 9; $b = $a; return str_pad('ab', $b, ''); }", 1},
		{"step after the call", "<?php function f() { $w = 9; $x = str_pad('ab', $w, ''); $w--; return $x; }", 1},
		{"step in a loop that ended", "<?php function f() { $w = 9; do { $x = str_pad('ab', $w, ''); } while (rand()); $w--; return $x; }", 1},
		{"extract after the call", "<?php $q = ''; echo str_pad('ab', 5, $q); extract(['q' => '*']);", 1},
		{"capture by value", "<?php function f() { $p = ''; $g = function () use ($p) {}; return str_pad('ab', 5, $p); }", 1},
		{"other function call before", "<?php function f() { $p = ''; trim('x'); return str_pad('ab', 5, $p); }", 1},
		{"variable variable argument stays unknown", "<?php function f() { $n = 'p'; return str_pad('ab', 5, $$n); }", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"StringPadEmptyPadding"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.name, len(got), tc.want, got)
		}
	}
}
