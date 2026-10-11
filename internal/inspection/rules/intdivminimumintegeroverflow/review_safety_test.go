package intdivminimumintegeroverflow

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
		{"plain dividend", "<?php function f() { $a = PHP_INT_MIN; return intdiv($a, -1); }", 1},
		{"plain divisor", "<?php function f() { $m = -1; return intdiv(PHP_INT_MIN, $m); }", 1},
		{"copied dividend", "<?php function f() { $b = PHP_INT_MIN; $a = $b; return intdiv($a, -1); }", 1},
		{"read by value", "<?php function f() { $a = PHP_INT_MIN; $x = abs($a); return intdiv($a, -1); }", 1},
		{"named by value", "<?php function f() { $a = PHP_INT_MIN; $x = abs(num: $a); return intdiv($a, -1); }", 1},
		{"variadic by value", "<?php function f() { $a = PHP_INT_MIN; $x = max(1, 2, $a); return intdiv($a, -1); }", 1},
		{"variadic tail by value", "<?php function f() { $a = PHP_INT_MIN; $x = sprintf('%d%d%d', 1, 2, $a); return intdiv($a, -1); }", 1},
		{"read in expression", "<?php function f() { $a = PHP_INT_MIN; $x = $a + 1; return intdiv($a, -1); }", 1},
		{"variadic tail by reference", "<?php function f(string $s) { $a = PHP_INT_MIN; sscanf($s, '%d%d', $x, $a); return intdiv($a, -1); }", 0},
		{"changed afterwards", "<?php function f() { $a = PHP_INT_MIN; $x = intdiv($a, -1); $a++; return $x; }", 1},
		{"top level", "<?php $a = PHP_INT_MIN; intdiv($a, -1);", 1},
		{"incremented dividend", "<?php function f() { $a = PHP_INT_MIN; $a++; return intdiv($a, -1); }", 0},
		{"decremented divisor", "<?php function f() { $a = PHP_INT_MIN; $m = -1; $m--; return intdiv($a, $m); }", 0},
		{"compound dividend", "<?php function f() { $a = PHP_INT_MIN; $a += 1; return intdiv($a, -1); }", 0},
		{"destructured dividend", "<?php function f(array $x) { $a = PHP_INT_MIN; [$a] = $x; return intdiv($a, -1); }", 0},
		{"foreach dividend", "<?php function f() { $a = PHP_INT_MIN; foreach ([1, 2] as $a) { intdiv($a, -1); } }", 0},
		{"foreach divisor", "<?php function f() { $m = -1; foreach ([1] as $m) {} return intdiv(PHP_INT_MIN, $m); }", 0},
		{"by-reference builtin", "<?php function f(string $s) { $a = PHP_INT_MIN; sscanf($s, '%d', $a); return intdiv($a, -1); }", 0},
		{"named by-reference builtin", "<?php function f() { $m = -1; preg_match(subject: 'x', pattern: '/x/', matches: $m); return intdiv(PHP_INT_MIN, $m); }", 0},
		{"unknown callee", "<?php function f() { $m = -1; unknown($m); return intdiv(PHP_INT_MIN, $m); }", 0},
		{"call-time reference", "<?php function f() { $m = -1; abs(&$m); return intdiv(PHP_INT_MIN, $m); }", 0},
		{"copied from changed source", "<?php function f() { $b = PHP_INT_MIN; $b++; $a = $b; return intdiv($a, -1); }", 0},
		{"changed later in loop", "<?php function f() { $a = PHP_INT_MIN; while (g()) { intdiv($a, -1); $a++; } }", 0},
		{"alias", "<?php function f() { $a = PHP_INT_MIN; $r = &$a; $r = 3; return intdiv($a, -1); }", 0},
		{"reference target", "<?php function f() { $b = 3; $a = PHP_INT_MIN; $a = &$b; return intdiv($a, -1); }", 0},
		{"global", "<?php function f() { $a = PHP_INT_MIN; global $a; return intdiv($a, -1); }", 0},
		{"static", "<?php function f() { $a = PHP_INT_MIN; static $a; return intdiv($a, -1); }", 0},
		{"mutual copies", "<?php function f() { $a = PHP_INT_MIN; $b = $a; $a = $b; $a++; return intdiv($a, -1); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
