package nonvoidfunctionfallsthrough

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewNonVoidFunctionFallsThrough(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		// Exhaustive sibling branches leave no reachable exit.
		{"ordering complement", "function f(int $n): int { if ($n > 0) { return 1; } if ($n <= 0) { return -1; } }", 0},
		{"reverse ordering complement", "function f($n): int { if ($n < 0) { return 1; } if ($n >= 0) { return -1; } }", 0},
		{"mirrored literal", "function f($n): int { if (0 < $n) { return 1; } if ($n <= 0) { return -1; } }", 0},
		{"mirrored variables", "function f($a, $b): int { if ($a < $b) { return 1; } if ($b <= $a) { return -1; } }", 0},
		{"mirrored variables greater", "function f($a, $b): int { if ($a > $b) { return 1; } if ($b >= $a) { return -1; } }", 0},
		{"mirrored equality", "function f($a, $b): int { if ($a == $b) { return 1; } if ($b != $a) { return -1; } }", 0},
		{"loose complement", "function f($n): int { if ($n == 1) { return 1; } if ($n != 1) { return -1; } }", 0},
		{"strict complement", "function f($n): int { if ($n === null) { return 1; } if ($n !== null) { return -1; } }", 0},
		{"parenthesized", "function f($n): int { if (($n) === (null)) { return 1; } if ((($n) !== null)) { return -1; } }", 0},
		{"bool parameter", "function f(bool $b): int { if ($b === true) { return 1; } if ($b === false) { return 0; } }", 0},
		{"bool parameter reversed", "function f(BOOL $b): int { if (false === $b) { return 1; } if ($b) { return 0; } }", 0},
		{"bool parameter not identical", "function f(bool $b): int { if ($b !== true) { return 1; } if ($b !== false) { return 0; } }", 0},
		{"loose bool literal", "function f($b): int { if ($b == true) { return 1; } if ($b == false) { return 0; } }", 0},
		{"loose bool then truthiness", "function f($b): int { if ($b != false) { return 1; } if ($b == false) { return 0; } }", 0},
		{"throwing branches", "function f($n): int { if ($n > 0) { throw new E(); } if ($n <= 0) { return 0; } }", 0},
		{"inert statement between", "function f($n): int { if ($n > 0) { return 1; } echo $n; ; if ($n <= 0) { return -1; } }", 0},
		{"nested complement", "function f($n): int { if ($n > 0) { return 1; } else { if ($n <= 0) { return -1; } } }", 0},
		{"else terminates", "function f($n): int { if ($n > 0) { echo 1; } else { return 0; } if ($n > 0) { return 1; } }", 0},
		{"write in terminating branch", "function f($n): int { if ($n > 0) { $n = 0; return 1; } if ($n <= 0) { return -1; } }", 0},
		{"repeated test in dead branch", "function f($n): int { if ($n > 0) { return 1; } if ($n > 0) { echo 1; } else { return 2; } }", 0},

		// Not exhaustive: a value reaches the end of the body.
		{"gap at zero", "function f(int $n): int { if ($n > 0) { return 1; } if ($n < 0) { return -1; } }", 1},
		{"same test twice", "function f($n): int { if ($n > 0) { return 1; } if ($n > 0) { return -1; } }", 1},
		{"different literal", "function f($n): int { if ($n > 0) { return 1; } if ($n <= 1) { return -1; } }", 1},
		{"different variable", "function f($n, $m): int { if ($n > 0) { return 1; } if ($m <= 0) { return -1; } }", 1},
		{"different variable pair", "function f($a, $b, $c): int { if ($a < $b) { return 1; } if ($c <= $a) { return -1; } }", 1},
		{"untyped strict bool", "function f($b): int { if ($b === true) { return 1; } if ($b === false) { return 0; } }", 1},
		{"nullable bool", "function f(?bool $b): int { if ($b === true) { return 1; } if ($b === false) { return 0; } }", 1},
		{"variadic bool", "function f(bool ...$b): int { if ($b === true) { return 1; } if ($b === false) { return 0; } }", 1},
		{"local strict bool", "function f(): int { if ($b === true) { return 1; } if ($b === false) { return 0; } }", 1},
		{"ordering against bool", "function f($b): int { if ($b > true) { return 1; } if ($b > false) { return 0; } }", 1},
		{"truthiness against comparison", "function f($b): int { if ($b) { return 1; } if ($b == 0) { return 0; } }", 1},
		{"write between", "function f($n, $m): int { if ($n > 0) { return 1; } $n = $m; if ($n <= 0) { return -1; } }", 1},
		{"increment between", "function f($n): int { if ($n > 0) { return 1; } $n++; if ($n <= 0) { return -1; } }", 1},
		{"write in open branch", "function f($n, $z): int { if ($n > 0) { return 1; } if ($z) { $n = 5; } if ($n <= 0) { return -1; } }", 1},
		{"write in surviving then", "function f($n, $m): int { if ($n > 0) { $n = $m; } else { return 0; } if ($n > 0) { return 1; } }", 1},
		{"write in surviving else", "function f($n, $m): int { if ($n > 0) { return 1; } else { $n = $m; } if ($n <= 0) { return -1; } }", 1},
		{"write in condition", "function f($n, $m): int { if ($n > 0) { return 1; } if ($n++ > $m) { return 2; } if ($n <= 0) { return -1; } }", 1},
		{"first branch open", "function f($n): int { if ($n > 0) { echo 1; } if ($n <= 0) { return -1; } }", 1},
		{"complement keeps else", "function f($n): int { if ($n > 0) { return 1; } if ($n <= 0) { echo 1; } else { return 2; } }", 1},
		{"repeated test without else", "function f($n): int { if ($n > 0) { return 1; } if ($n > 0) { echo 1; } }", 1},
		{"dynamic variable", "function f($n): int { if ($$n > 0) { return 1; } if ($$n <= 0) { return -1; } }", 1},
		{"property operands", "function f($o): int { if ($o->p > 0) { return 1; } if ($o->p <= 0) { return -1; } }", 1},
		{"property right operand", "function f($n, $o): int { if ($n > $o->p) { return 1; } if ($n <= $o->p) { return -1; } }", 1},
		{"unbraced nested if", "function f($n, $m): int { if ($n) if ($m) return 1; }", 1},
		{"known operands stay unproven", "function f($n): int { if ($n > 0) { return 1; } if (1 > 2) { return 2; } }", 0},

		// Expressions that cannot complete.
		{"match default throws", "function f($e): int { match (true) { default => throw $e }; }", 0},
		{"match every arm throws", "function f($e, $x): int { match ($x) { 1 => throw $e, 2 => (throw $e) }; }", 0},
		{"match subject throws", "function f($e): int { match (throw $e) { 1 => 2 }; }", 0},
		{"assigned throw", "function f($e): int { $x = throw $e; }", 0},
		{"assigned throwing match", "function f($e): int { $x = match (true) { default => throw $e }; }", 0},
		{"match arm completes", "function f($e, $x): int { match ($x) { 1 => throw $e, default => 2 }; }", 1},
		{"coalescing assignment may skip throw", "function f($e, $x): int { $x ??= throw $e; }", 1},
		{"short-circuit may skip throw", "function f($e, $x): int { $x || throw $e; }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
