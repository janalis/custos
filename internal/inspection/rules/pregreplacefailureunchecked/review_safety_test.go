package pregreplacefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewNullGuards(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const produce = "<?php function f($input, $x) { $t = preg_replace('/z/u', '', $input); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"unguarded", "return trim($t);", 1},
		{"strict ternary else", "return $t === null ? '' : trim($t);", 0},
		{"strict ternary then", "return null !== ($t) ? trim(($t)) : '';", 0},
		{"wrong ternary arm", "return $t === null ? trim($t) : '';", 1},
		{"short ternary", "return $x ?: trim($t);", 1},
		{"is_null return", "if (is_null($t)) { return ''; } return trim($t);", 0},
		{"is_null branch", "if (is_null($t)) { return trim($t); } return '';", 1},
		{"is_null without argument", "if (is_null()) { return ''; } return trim($t);", 1},
		{"is_string branch", "if (is_string($t)) { return trim($t); } return '';", 0},
		{"negated is_string return", "if (!is_string($t)) { return ''; } return trim($t);", 0},
		{"negated is_string branch", "if (!is_string($t)) { return trim($t); } return '';", 1},
		{"other predicate", "if (is_array($t)) { return ''; } return trim($t);", 1},
		{"bitwise not", "if (~$x) { return ''; } return trim($t);", 1},
		{"and operand", "return $t !== null && strlen($t) > 3;", 0},
		{"word and operand", "return ($t !== null and strlen($t) > 3);", 0},
		{"and left operand", "return strlen($t) > 3 && $t !== null;", 1},
		{"or operand", "if ($t === null || strlen($t) === 0) { return ''; } return $t;", 0},
		{"word or operand", "return ($t === null or strlen($t) === 0);", 0},
		{"wrong or operand", "return $t !== null || strlen($t) === 0;", 1},
		{"arithmetic operand", "return $t !== null + strlen($t);", 1},
		{"and condition", "if (is_string($t) && $x) { return trim($t); } return '';", 0},
		{"or condition", "if (is_string($t) || $x) { return trim($t); } return '';", 1},
		{"or return", "if ($x || $t === null) { return ''; } return trim($t);", 0},
		{"and return", "if ($x && $t === null) { return ''; } return trim($t);", 1},
		{"isset ternary", "return isset($t) ? trim($t) : '';", 0},
		{"isset several", "return isset($x, $t) ? trim($t) : '';", 0},
		{"isset other", "return isset($x) ? trim($t) : '';", 1},
		{"negated isset", "return !isset($t) ? trim($t) : '';", 1},
		{"loose comparison", "if ($t == null) { return ''; } return trim($t);", 1},
		{"truthiness", "if ($t) { return trim($t); } return '';", 1},
		{"other constant", "if ($t === false) { return ''; } return trim($t);", 1},
		{"other variable", "if ($x === null) { return ''; } return trim($t);", 1},
		{"earlier guard", "if ($t === null) { return ''; } echo $x; return trim($t);", 0},
		{"nonterminating guard", "if ($t === null) { echo $x; } return trim($t);", 1},
		{"guard with else", "if (is_null($t)) { return ''; } else { echo $x; } return trim($t);", 0},
		{"guard with elseif", "if (is_null($t)) { return ''; } elseif ($x) { echo $x; } return trim($t);", 0},
		{"elseif is_string body", "if ($x) { return ''; } elseif (is_string($t)) { return trim($t); } return '';", 0},
		{"elseif is_null body", "if ($x) { return ''; } elseif (is_null($t)) { return trim($t); } return '';", 1},
		{"elseif body", "if ($x) { return ''; } elseif ($t !== null) { return trim($t); } return '';", 0},
		{"elseif after null test", "if ($t === null) { return ''; } elseif ($x) { return trim($t); } return '';", 0},
		{"elseif condition", "if ($t === null) { return ''; } elseif (trim($t)) { return ''; } return '';", 0},
		{"else after null test", "if ($t === null) { return ''; } else { return trim($t); }", 0},
		{"else after other test", "if ($x) { return ''; } else { return trim($t); }", 1},
		{"unbraced body", "if ($t !== null) return trim($t); return '';", 0},
		{"reassigned after guard", "if ($t === null) { return ''; } $t = preg_replace('/y/u', '', $t); return trim($t);", 1},
		{"closure has own scope", "if ($t === null) { return ''; } return function ($input) { $t = preg_replace('/z/u', '', $input); return trim($t); };", 1},
		{"assignment guard", "if (null !== $u = preg_replace('/y/u', '', $input)) { return trim($u); } return '';", 0},
		{"assignment guard return", "if (($u = preg_replace('/y/u', '', $input)) === null) { return ''; } return trim($u);", 0},
		{"assignment to other variable", "if (null !== $u = preg_replace('/y/u', '', $input)) { return trim($t); } return '';", 1},
		{"compound assignment guard", "if (null !== $t .= $x) { return trim($t); } return '';", 0},
		{"assignment guard then reassigned", "if (null !== $u = preg_replace('/y/u', '', $input)) { $u = preg_replace('/y/u', '', $input); return trim($u); } return '';", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(produce+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
