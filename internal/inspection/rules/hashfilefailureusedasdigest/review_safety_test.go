package hashfilefailureusedasdigest

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
	const head = "<?php function f($p, $m, $o) { $d = hash_file('sha256', $p); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"unguarded", "file_put_contents($m, $d); }", 1},
		{"later sibling exit", "if ($d === false) { return; } $l = basename($p); file_put_contents($m, $d); }", 0},
		{"truthy exit", "if (!$d) { return; } file_put_contents($m, $d); }", 0},
		{"empty exit", "if (empty($d)) { return; } $l = 1; file_put_contents($m, $d); }", 0},
		{"loose exit", "if ($d == false) { throw new RuntimeException(); } file_put_contents($m, $d); }", 0},
		{"yoda exit", "if (false === ($d)) { return; } echo 1; file_put_contents($m, $d); }", 0},
		{"or exit", "if (!$d || $m === '') { return; } file_put_contents($m, $d); }", 0},
		{"and exit keeps false", "if (!$d && $m === '') { return; } file_put_contents($m, $d); }", 1},
		{"and exit both reject", "if (!$d && empty($d)) { return; } file_put_contents($m, $d); }", 0},
		{"exit with else", "if (!$d) { return; } else { echo 1; } file_put_contents($m, $d); }", 0},
		{"exit with elseif", "if (!$d) { return; } elseif ($o) { echo 1; } file_put_contents($m, $d); }", 0},
		{"non terminating guard", "if (!$d) { echo 1; } file_put_contents($m, $d); }", 1},
		{"other variable", "if (!$o) { return; } file_put_contents($m, $d); }", 1},
		{"other comparison", "if ($d === null) { return; } file_put_contents($m, $d); }", 1},
		{"non boolean condition", "if ($d . $o) { file_put_contents($m, $d); } }", 1},
		{"negated other operator", "if (-$d) { file_put_contents($m, $d); } }", 1},
		{"reassigned after guard", "if (!$d) { return; } $d = hash_file('md5', $p); file_put_contents($m, $d); }", 1},
		{"guard in outer block", "if (!$d) { return; } if ($o) { file_put_contents($m, $d); } }", 0},
		{"truthy branch", "if ($d) { file_put_contents($m, $d); } }", 0},
		{"loose branch", "if ($d != false) { file_put_contents($m, $d); } }", 0},
		{"and branch", "if ($o && $d) { file_put_contents($m, $d); } }", 0},
		{"or branch", "if ($o || $d) { file_put_contents($m, $d); } }", 1},
		{"or branch both reject", "if ($d || !empty($d)) { file_put_contents($m, $d); } }", 0},
		{"else branch", "if (!$d) { echo 1; } else { file_put_contents($m, $d); } }", 0},
		{"else branch unguarded", "if ($d) { echo 1; } else { file_put_contents($m, $d); } }", 1},
		{"failure branch", "if (!$d) { file_put_contents($m, $d); } }", 1},
		{"short circuit and", "return $d !== false && $d === 'abc'; }", 0},
		{"short circuit keyword", "return ($d and 'abc' === $d); }", 0},
		{"short circuit or", "return $d === false || $d !== 'abc'; }", 0},
		{"short circuit wrong side", "return $d !== false || $d === 'abc'; }", 1},
		{"left operand", "return $d === 'abc' && $d !== false; }", 1},
		{"ternary then", "return $d ? $d === 'abc' : false; }", 0},
		{"ternary else", "return !$d ? false : $d === 'abc'; }", 0},
		{"ternary wrong arm", "return $d ? false : $d === 'abc'; }", 1},
		{"short ternary", "return $o ?: $d === 'abc'; }", 1},
		// A disjunction that held, or a conjunction that failed, with both operands
		// rejecting false is proven by the guard walk only.
		{"short circuit and both reject", "return ($d || !empty($d)) && $d === 'abc'; }", 0},
		{"short circuit or both reject", "return (!$d && empty($d)) || $d !== 'abc'; }", 0},
		{"ternary then both reject", "return ($d || !empty($d)) ? $d === 'abc' : false; }", 0},
		{"ternary else both reject", "return (!$d && empty($d)) ? false : $d === 'abc'; }", 0},
		{"else branch both reject", "if (!$d && empty($d)) { echo 1; } else { file_put_contents($m, $d); } }", 0},
		{"and branch both reject", "if ($o && ($d || !empty($d))) { file_put_contents($m, $d); } }", 0},
		{"else branch one rejects", "if (!$d && $o) { echo 1; } else { file_put_contents($m, $d); } }", 1},
		{"inline call", "} function g($p, $m) { if ($m) { file_put_contents($m, hash_file('sha256', $p)); } }", 1},
		{"closure boundary", "if (!$d) { return; } return function () use ($d, $m) { file_put_contents($m, $d); }; }", 0},
		{"missing data", "file_put_contents($m); }", 0},
		{"loose digest comparison", "return $d == 'abc'; }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(head+tc.body), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			for _, g := range got {
				if len(g.Fixes) != 0 {
					t.Fatal("unexpected fix")
				}
			}
		})
	}
}
