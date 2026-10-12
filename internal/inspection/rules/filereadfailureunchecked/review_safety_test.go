package filereadfailureunchecked

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
	const (
		decl = "<?php function stop(): never { throw new Exception('x'); } class K { static function stop(): never { exit; } } "
		read = "$c = file_get_contents($f); "
		use  = " return strlen($c); }"
	)
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		// Alternatives that never return.
		{"or die", "function f($f) { $c = file_get_contents($f) or die('x');" + use, 0},
		{"or exit parens", "function f($f) { ($c = (file_get_contents($f))) || exit(1);" + use, 0},
		{"or throw", "function f($f) { $c = file_get_contents($f) or throw new Exception('x');" + use, 0},
		{"or never function", "function f($f) { $c = file_get_contents($f) or stop();" + use, 0},
		{"or never static", "function f($f) { $c = file_get_contents($f) or K::stop();" + use, 0},
		{"or returning call", "function f($f) { $c = file_get_contents($f) or print('x');" + use, 1},
		{"or unknown call", "function f($f) { $c = file_get_contents($f) or unknown();" + use, 1},
		{"and die", "function f($f) { $c = file_get_contents($f) and die('x');" + use, 1},
		{"use inside alternative", "function f($f) { $c = file_get_contents($f) or die(strlen($c)); }", 1},
		// Loose comparisons with a falsy constant.
		{"loose false", "function f($f) { " + read + "if ($c == false) { return 0; }" + use, 0},
		{"yoda false", "function f($f) { " + read + "if (FALSE == $c) { return 0; }" + use, 0},
		{"loose null", "function f($f) { " + read + "if ($c == null) { return 0; }" + use, 0},
		{"loose empty string", "function f($f) { " + read + "if ($c == '') { return 0; }" + use, 0},
		{"loose zero", "function f($f) { " + read + "if ($c == 0) { return 0; }" + use, 0},
		{"loose true", "function f($f) { " + read + "if ($c == true) { return 0; }" + use, 1},
		{"loose text", "function f($f) { " + read + "if ($c == 'x') { return 0; }" + use, 1},
		{"loose unknown", "function f($f, $o) { " + read + "if ($c == $o) { return 0; }" + use, 1},
		{"strict empty string", "function f($f) { " + read + "if ($c === '') { return 0; }" + use, 1},
		{"other variable", "function f($f, $d) { " + read + "if ($d == false) { return 0; }" + use, 1},
		{"other expression", "function f($f) { " + read + "if (g() == false) { return 0; }" + use, 1},
		{"other read", "function f($f) { " + read + "if (($d = file_get_contents($f)) == false) { return 0; }" + use, 1},
		{"read after guard", "function f($f) { " + read + "if ($c == false) { return 0; } $c = file_get_contents($f);" + use, 1},
		{"guard does not exit", "function f($f) { " + read + "if ($c == false) { echo 1; }" + use, 1},
		// A loose comparison with a falsy constant is read by the rule itself.
		{"loose empty string branch", "function f($f) { " + read + "if ($c != '') {" + use + " }", 0},
		{"loose empty string operand", "function f($f) { " + read + "return $c != '' && strlen($c); }", 0},
		{"loose empty string arm", "function f($f) { " + read + "return $c == '' ? 0 : strlen($c); }", 0},
		{"loose empty string else", "function f($f) { " + read + "if ($c == '') { echo 1; } else {" + use + " }", 0},
		{"loose empty string unbraced never", "function f($f) { " + read + "if ($c == '') stop();" + use, 0},
		{"loose empty string wrong branch", "function f($f) { " + read + "if ($c == '') {" + use + " }", 1},
		{"loose non-empty string", "function f($f) { " + read + "if ($c != 'x') {" + use + " }", 1},
		// The use is only reached when the comparison failed, whatever the else branch does.
		{"guard with else", "function f($f) { " + read + "if ($c == false) { return 0; } else { echo 1; }" + use, 0},
		{"conjunction exit", "function f($f, $x) { " + read + "if ($c == false && $x) { return 0; }" + use, 1},
		{"disjunction exit", "function f($f, $x) { " + read + "if ($x || $c == false) { return 0; }" + use, 0},
		{"negated exit", "function f($f) { " + read + "if (!($c != false)) { return 0; }" + use, 0},
		{"assigned in condition", "function f($f) { if (($c = file_get_contents($f)) == false) { return 0; }" + use, 0},
		{"top level", read + "if ($c == false) { exit(1); } echo strlen($c);", 0},
		{"unbraced use", "function f($f, $x) { " + read + "if ($c == false) { return 0; } if ($x) return strlen($c); }", 0},
		// Branches selected by the guard.
		{"then branch", "function f($f) { " + read + "if ($c != false) { return strlen($c); } }", 0},
		{"then branch conjunction", "function f($f, $x) { " + read + "if ($x && $c <> false) { return strlen($c); } }", 0},
		{"then branch disjunction", "function f($f, $x) { " + read + "if ($x || $c != false) { return strlen($c); } }", 1},
		{"then branch of equality", "function f($f) { " + read + "if ($c == false) { return strlen($c); } }", 1},
		{"else branch", "function f($f) { " + read + "if ($c == false) { return 0; } else { return strlen($c); } }", 0},
		{"elseif branch", "function f($f, $x) { " + read + "if ($c == false) { return 0; } elseif ($x) { return strlen($c); } }", 0},
		{"else of inequality", "function f($f) { " + read + "if ($c != false) { return 0; } else { return strlen($c); } }", 1},
		{"ternary else", "function f($f) { " + read + "return $c == false ? 0 : strlen($c); }", 0},
		{"ternary then", "function f($f) { " + read + "return $c != false ? strlen($c) : 0; }", 0},
		{"ternary wrong side", "function f($f) { " + read + "return $c == false ? strlen($c) : 0; }", 1},
		{"ternary condition", "function f($f) { " + read + "return strlen($c) ? 1 : 0; }", 1},
		{"short circuit and", "function f($f) { " + read + "return $c != false && strlen($c); }", 0},
		{"short circuit or", "function f($f) { " + read + "return $c == false || strlen($c); }", 0},
		{"short circuit wrong", "function f($f) { " + read + "return $c != false || strlen($c); }", 1},
		{"left operand", "function f($f) { " + read + "return strlen($c) && $c != false; }", 1},
		// Branches ending in a call declared never.
		{"never function", "function f($f) { " + read + "if ($c === false) { stop(); }" + use, 0},
		{"never after statement", "function f($f) { " + read + "if (false === $c) { echo 1; stop(); }" + use, 0},
		{"never unbraced", "function f($f) { " + read + "if (!$c) stop();" + use, 0},
		{"never static", "function f($f) { " + read + "if ($c === false) { K::stop(); }" + use, 0},
		{"never not string", "function f($f) { " + read + "if (!is_string($c)) { stop(); }" + use, 0},
		{"never is string exits", "function f($f) { " + read + "if (is_string($c)) { stop(); }" + use, 1},
		// Execution continues only for an array, which is not the failure value.
		{"never other predicate", "function f($f) { " + read + "if (!is_array($c)) { stop(); }" + use, 0},
		{"never empty", "function f($f) { " + read + "if (empty($c)) { stop(); }" + use, 0},
		{"never not empty", "function f($f) { " + read + "if (!empty($c)) { stop(); }" + use, 1},
		{"never truthy exits", "function f($f) { " + read + "if ($c) { stop(); }" + use, 1},
		{"never strict true", "function f($f) { " + read + "if ($c === true) { stop(); }" + use, 1},
		{"never strict inequality", "function f($f) { " + read + "if ($c !== false) { stop(); }" + use, 1},
		{"never arithmetic", "function f($f) { " + read + "if ($c + 1) { stop(); }" + use, 1},
		{"returning helper", "function f($f) { " + read + "if ($c === false) { unknown(); }" + use, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := tc.source
			if len(src) < 5 || src[:5] != "<?php" {
				src = decl + src
			}
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(src), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
