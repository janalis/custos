package nonexhaustiveenummatch_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/nonexhaustiveenummatch"
	"custos/internal/php/syntax"
)

// TestReviewNarrowingAndAliases pins the review findings: comparisons that
// exclude the missing case on every path silence the report, an unproven
// subject keeps it, and constants aliasing a case count as that case.
func TestReviewNarrowingAndAliases(t *testing.T) {
	const enum = "<?php\nenum E { case A; case B; const X = self::A; }\nfunction byRef(E &$r): void {}\n"
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"early return excludes case", "function f(E $e) { if ($e === E::A) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"enclosing negated guard", "function f(E $e) { if ($e !== E::A) { return match($e) { E::B => 1 }; } return 0; }", 0},
		{"ternary else branch", "function f(E $e) { return $e === E::B ? 0 : match($e) { E::A => 1 }; }", 0},
		{"ternary then branch pins case", "function f(E $e) { return $e === E::B ? match($e) { E::B => 1 } : 0; }", 0},
		{"alias arm covers its case", "function f(E $e) { return match($e) { E::X => 1, E::B => 2 }; }", 0},
		{"alias arm leaves other case", "function f(E $e) { return match($e) { E::X => 1 }; }", 1},
		{"alias in guard", "function f(E $e) { if ($e === E::X) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"guard on wrong case", "function f(E $e) { if ($e === E::B) { return 0; } return match($e) { E::B => 1 }; }", 1},
		{"guard does not exit", "function f(E $e) { if ($e === E::A) { echo 1; } return match($e) { E::B => 1 }; }", 1},
		{"guard on other variable", "function f(E $e, E $o) { if ($o === E::A) { return 0; } return match($e) { E::B => 1 }; }", 1},
		{"subject reassigned", "function f(E $e, E $o) { if ($e === E::A) { return 0; } $e = $o; return match($e) { E::B => 1 }; }", 1},
		{"subject type unknown after reference call", "function f(E $e) { if ($e === E::A) { return 0; } byRef($e); return match($e) { E::B => 1 }; }", 0},
		{"by-reference parameter", "function f(E &$e) { if ($e === E::A) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"by-reference parameter across a call", "function f(E &$e) { if ($e === E::A) { return 0; } g(); return match($e) { E::B => 1 }; }", 1},
		{"reference alias write", "function f(E $e) { if ($e === E::A) { return 0; } $r = &$e; $r = E::A; return match($e) { E::B => 1 }; }", 1},
		{"reference alias written between", "function f(E $e) { $r = &$e; if ($e === E::A) { return 0; } $r = E::A; return match($e) { E::B => 1 }; }", 1},
		{"reference alias untouched", "function f(E $e) { $r = &$e; if ($e === E::A) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"eval", "function f(E $e) { if ($e === E::A) { return 0; } eval('$e = E::A;'); return match($e) { E::B => 1 }; }", 0},
		{"jump target", "function f(E $e) { if ($e === E::A) { return 0; } l: echo match($e) { E::B => 1 }; goto l; }", 1},
		{"loop leaves on one case", "function f(E $e) { while ($e !== E::B) { $e = E::B; } return match($e) { E::B => 1 }; }", 0},
		{"either case", "function f(E $e) { if ($e === E::A || $e === E::B) { return match($e) { E::A => 1 }; } }", 1},
		{"assigned case", "function f() { $e = E::B; return match($e) { E::B => 1 }; }", 0},
		{"alias assigned", "function f() { $e = E::X; return match($e) { E::A => 1 }; }", 0},
		{"enum function", "function g(): E {} function f() { $e = g(); if ($e === E::A) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"switch label", "function f(E $e) { switch ($e) { case E::A: return 0; default: return match($e) { E::B => 1 }; } }", 1},
		{"enclosing match default", "function f(E $e) { return match($e) { E::A => 0, default => match($e) { E::B => 1 } }; }", 0},
		{"enclosing match true default", "function f(E $e) { return match(true) { $e === E::A => 0, default => match($e) { E::B => 1 } }; }", 0},
		{"assert", "function f(E $e) { assert($e !== E::A); return match($e) { E::B => 1 }; }", 1},
		{"in_array", "function f(E $e) { if (in_array($e, [E::A], true)) { return 0; } return match($e) { E::B => 1 }; }", 1},
		{"saved comparison", "function f(E $e) { $ok = $e === E::A; if ($ok) { return 0; } return match($e) { E::B => 1 }; }", 0},
		{"comparison compared with false", "function f(E $e) { if (($e === E::A) === false) { return match($e) { E::B => 1 }; } }", 0},
		{"enum method", "enum F { case P; case Q; function g() { if ($this === self::P) { return 0; } return match($this) { self::Q => 1 }; } }", 0},
		{"enum method wrong case", "enum F { case P; case Q; function g() { if ($this === self::P) { return 0; } return match($this) { self::P => 1 }; } }", 1},
		{"three cases one left", "enum T { case P; case Q; case R; } function f(T $t) { if ($t === T::P || $t === T::Q) { return 0; } return match($t) { T::R => 1 }; }", 0},
		{"three cases two left", "enum T { case P; case Q; case R; } function f(T $t) { if ($t === T::P) { return 0; } return match($t) { T::R => 1 }; }", 1},
		{"loose comparison", "function f(E $e) { if ($e == E::A) { return 0; } return match($e) { E::B => 1 }; }", 1},
		{"top-level guard", "if ($e === E::A) { exit; } $e = E::B; echo match($e) { E::A => 1 };", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			findings := e.Analyze(syntax.Parse("test.php", []byte(enum+tc.source), syntax.Options{}))
			if len(findings) != tc.count {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.count, findings)
			}
		})
	}
}
