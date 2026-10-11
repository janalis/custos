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
		{"subject passed by reference", "function f(E $e) { if ($e === E::A) { return 0; } byRef($e); return match($e) { E::B => 1 }; }", 1},
		{"by-reference parameter", "function f(E &$e) { if ($e === E::A) { return 0; } return match($e) { E::B => 1 }; }", 1},
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
