package xpathqueryfailuredereferenced

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
		name, body string
		findings   int
	}{
		{"falsy exit", "$n = $x->query($e); if (!$n) { return 0; } return $n->length;", 0},
		{"truthy branch", "$n = $x->query($e); if ($n) { return $n->length; } return 0;", 0},
		{"instanceof branch", "$n = $x->query($e); if ($n instanceof DOMNodeList) { return $n->length; } return 0;", 0},
		{"instanceof exit", "$n = $x->query($e); if (!$n instanceof DOMNodeList) { return 0; } return $n->length;", 0},
		{"conjunction", "$n = $x->query($e); if ($n !== false && $n->length > 0) { return 1; } return 0;", 0},
		{"conjunction body", "$n = $x->query($e); if ($e !== '' && $n) { return $n->length; } return 0;", 0},
		{"disjunction exit", "$n = $x->query($e); if ($e === '' || $n === false) { return 0; } return $n->length;", 0},
		{"disjunction operand", "$n = $x->query($e); return (int) ($n === false || $n->length === 0);", 0},
		{"ternary", "$n = $x->query($e); return $n !== false ? $n->length : 0;", 0},
		{"ternary else", "$n = $x->query($e); return $n === false ? 0 : $n->length;", 0},
		{"inline assignment", "if (($n = $x->query($e)) === false) { return 0; } return $n->length;", 0},
		{"inline truthy", "if ($n = $x->query($e)) { return $n->length; } return 0;", 0},
		{"later statement", "$n = $x->query($e); if ($n === false) { return 0; } $seen = 0; return $n->length + $seen;", 0},
		{"throwing guard", "$n = $x->query($e); if ($n === false) { throw new RuntimeException('bad'); } $seen = 0; return $n->length + $seen;", 0},
		{"else branch", "$n = $x->query($e); if ($n === false) { return 0; } else { return $n->length; }", 0},
		{"reversed operands", "$n = $x->query($e); if (false === ($n)) { return 0; } return $n->length;", 0},
		{"qualified constant", "$n = $x->query($e); if (\\false !== $n) { return $n->length; } return 0;", 0},
		{"loose inequality", "$n = $x->query($e); return $n != false ? $n->length : 0;", 0},
		{"parenthesised receiver", "$n = $x->query($e); if (!$n) { return 0; } return ($n)->length;", 0},
		{"elseif branch", "$n = $x->query($e); if ($e === '') { return 0; } elseif ($n) { return $n->length; } return 0;", 0},
		{"after failed elseif", "$n = $x->query($e); if ($e === '') { return 0; } elseif (!$n) { return 0; } elseif ($e === '/') { return $n->length; } return 0;", 0},
		{"else after failed if", "$n = $x->query($e); if (!$n) { $e = ''; } elseif ($n->length > 1) { return 2; } else { return $n->length; } return 0;", 0},
		{"else after failed elseif", "$n = $x->query($e); if ($e === '') { return 0; } elseif ($n === false) { $e = ''; } else { return $n->length; } return 0;", 0},
		{"unbraced exit", "$n = $x->query($e); if (!$n) return 0; $seen = 1; if ($seen) return $n->length; return 0;", 0},
		{"nested use", "$n = $x->query($e); if (!$n) { return 0; } foreach ([1] as $i) { if ($i) { return $n->item($i) ? 1 : 0; } } return 0;", 0},
		{"loose false", "$n = $x->query($e); if ($n == false) { return 0; } return $n->length;", 0},

		{"null coalescing binding", "$n = null; if (($n ??= $x->query($e)) === false) { return 0; } return $n->length;", 0},

		{"unguarded", "$n = $x->query($e); return $n->length;", 1},
		{"direct", "return $x->query($e)->length;", 1},
		{"method", "$n = $x->query($e); return $n->item(0) === null ? 0 : 1;", 1},
		{"wrong branch", "$n = $x->query($e); if (!$n) { return $n->length; } return 0;", 1},
		{"failure branch", "$n = $x->query($e); if ($n === false) { return $n->length; } return 0;", 1},
		{"ternary failure arm", "$n = $x->query($e); return $n === false ? $n->length : 0;", 1},
		{"disjunction body", "$n = $x->query($e); if ($e === '' || $n) { return $n->length; } return 0;", 1},
		{"conjunction exit", "$n = $x->query($e); if ($e === '' && !$n) { return 0; } return $n->length;", 1},
		{"non terminating guard", "$n = $x->query($e); if ($n === false) { $e = ''; } return $n->length;", 1},
		{"other variable", "$n = $x->query($e); $m = $x->query($e); if ($m === false) { return 0; } return $n->length;", 1},
		{"requeried after guard", "$n = $x->query($e); if ($n === false) { return 0; } $n = $x->query($e . '/a'); return $n->length;", 1},
		{"merged producers stay unknown", "$n = $x->query($e); if ($e === '') { $n = $x->query('//a'); } return $n->length;", 0},
		// Escaped, dynamic and compound-written results have no proven producer.
		{"reference escape", "$n = $x->query($e); if ($n === false) { return 0; } reset_it($n); return $n->length;", 0},
		{"dynamic variable", "$k = 'n'; $n = $x->query($e); if ($$k === false) { return 0; } return $$k->length;", 0},
		{"negated instanceof branch", "$n = $x->query($e); if (!$n instanceof DOMNodeList) { return $n->length; } return 0;", 1},
		{"instanceof other variable", "$n = $x->query($e); if ($x instanceof DOMXPath) { return $n->length; } return 0;", 1},
		{"short ternary", "$n = $x->query($e); return (int) ($n ?: $n->length);", 1},
		{"own condition", "$n = $x->query($e); if ($n->length > 0) { return 1; } return 0;", 1},
		{"elseif condition", "$n = $x->query($e); if ($e === '') { return 0; } elseif ($n->length > 0) { return 1; } return 0;", 1},
		{"elseif unrelated", "$n = $x->query($e); if ($e === '') { return 0; } elseif ($e === '/') { return $n->length; } return 0;", 1},
		{"else unrelated", "$n = $x->query($e); if ($e === '') { return 0; } elseif ($e === '/') { return 1; } else { return $n->length; }", 1},
		{"compound assignment", "$n = $x->query($e); if (($n .= '') === false) { return 0; } return $n->length;", 0},
		{"unrelated comparison", "$n = $x->query($e); if ($n > 0) { return $n->length; } return 0;", 1},
		{"guard with else", "$n = $x->query($e); if ($n === false) { return 0; } else { $e = ''; } return $n->length;", 0},
		{"guard in closure", "$n = $x->query($e); $g = function () use ($n) { if ($n === false) { return 0; } return 1; }; return $n->length + $g();", 1},
		{"guard in loop", "$n = $x->query($e); foreach ([1] as $i) { if ($n === false) { continue; } } return $n->length;", 1},
		{"null guard", "$n = $x->query($e); if ($n === null) { return 0; } return $n->length;", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php function f(DOMXPath $x, string $e): int { "+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
