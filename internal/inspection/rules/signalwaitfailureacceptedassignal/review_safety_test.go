package signalwaitfailureacceptedassignal

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
	const wait = "$s = pcntl_sigwaitinfo($set); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		// Assigned in the condition: the call may sit anywhere in the branch.
		{"later statement", "if ($s = pcntl_sigwaitinfo($set)) { echo 1; handle($s); }", 1},
		{"alternative syntax", "if ($s = pcntl_sigwaitinfo($set)): echo 1; handle($s); endif;", 1},
		{"plain read first", "if ($s = pcntl_sigwaitinfo($set)) { $t = [$s]; echo $s; handle($s); }", 1},
		{"value capture first", "if ($s = pcntl_sigwaitinfo($set)) { $f = function () use ($s) {}; handle($s); }", 1},
		{"earlier plain call", "if ($s = pcntl_sigwaitinfo($set)) { handle(wrap($s)); }", 1},
		{"write after call", "if ($s = pcntl_sigwaitinfo($set)) { handle($s); $s = 0; }", 1},
		{"property target first", "if ($this->s = pcntl_sigwaitinfo($set)) { handle($this->s); }", 1},
		{"property target later", "if ($this->s = pcntl_sigwaitinfo($set)) { echo 1; handle($this->s); }", 0},
		{"dynamic target", "if ($$n = pcntl_sigwaitinfo($set)) { echo 1; handle($$n); }", 0},
		{"rebound", "if ($s = pcntl_sigwaitinfo($set)) { $s = 5; handle($s); }", 0},
		{"incremented", "if ($s = pcntl_sigwaitinfo($set)) { $s++; handle($s); }", 0},
		{"element write", "if ($s = pcntl_sigwaitinfo($set)) { $s[0] = 1; handle($s); }", 0},
		{"reference alias", "if ($s = pcntl_sigwaitinfo($set)) { $r = &$s; $r = 0; handle($s); }", 0},
		{"reference item", "if ($s = pcntl_sigwaitinfo($set)) { $r = [&$s]; fix($r); handle($s); }", 0},
		{"reference capture", "if ($s = pcntl_sigwaitinfo($set)) { $f = function () use (&$s) { $s = 0; }; $f(); handle($s); }", 0},
		{"method argument", "if ($s = pcntl_sigwaitinfo($set)) { $o->fix($s); handle($s); }", 0},
		{"static argument", "if ($s = pcntl_sigwaitinfo($set)) { Fixer::fix($s); handle($s); }", 0},
		{"variable variable", "if ($s = pcntl_sigwaitinfo($set)) { $$n = 0; handle($s); }", 0},
		{"extract", "if ($s = pcntl_sigwaitinfo($set)) { extract($vars); handle($s); }", 0},
		{"parse_str", "if ($s = pcntl_sigwaitinfo($set)) { parse_str($q, $out); eval($code); handle($s); }", 0},
		{"eval", "if ($s = pcntl_sigwaitinfo($set)) { eval($code); handle($s); }", 0},
		{"include", "if ($s = pcntl_sigwaitinfo($set)) { include $file; handle($s); }", 0},
		{"unpacked argument", "if ($s = pcntl_sigwaitinfo($set)) { handle(...$s); }", 0},
		{"other variable written", "if ($s = pcntl_sigwaitinfo($set)) { $t = 5; handle($s); }", 1},

		// Failure rejected before the call.
		{"earlier return", wait + "if ($s === -1) { return; } if ($s) { handle($s); }", 0},
		{"earlier return reversed", wait + "if (-1 == $s) return; if ($s) { handle($s); }", 0},
		{"earlier throw", wait + "if ($s < 0) { throw new E(); } if ($s) { handle($s); }", 0},
		{"earlier negated", wait + "if (!($s > 0)) { return; } if ($s) { handle($s); }", 0},
		{"earlier disjunction", wait + "if ($x || $s <= -1) { return; } if ($s) { handle($s); }", 0},
		{"earlier conjunction", wait + "if ($x && $s === -1) { return; } if ($s) { handle($s); }", 1},
		{"earlier non-leaving", wait + "if ($s === -1) { echo 1; } if ($s) { handle($s); }", 1},
		{"earlier other value", wait + "if ($t === -1) { return; } if ($s) { handle($s); }", 1},
		{"earlier other bound", wait + "if ($s === 0) { return; } if ($s) { handle($s); }", 1},
		{"earlier unknown bound", wait + "if ($s === $t) { return; } if ($s) { handle($s); }", 1},
		{"earlier arithmetic", wait + "if ($s + 1) { return; } if ($s) { handle($s); }", 1},
		{"earlier negated call", wait + "if (!$x) { return; } if ($s) { handle($s); }", 1},
		{"earlier bitwise not", wait + "if (~$s) { return; } if ($s) { handle($s); }", 1},
		{"guard before wait", "if ($s === -1) { return; } " + wait + "if ($s) { handle($s); }", 1},
		{"guard inside branch", "if ($s = pcntl_sigwaitinfo($set)) { if ($s === -1) { return; } handle($s); }", 0},
		{"guard after call", "if ($s = pcntl_sigwaitinfo($set)) { handle($s); if ($s === -1) { return; } }", 1},
		{"nested positive", wait + "if ($s) { if ($s > 0) { handle($s); } }", 0},
		{"nested not failure", wait + "if ($s) { if ($s !== -1) { handle($s); } }", 0},
		{"nested at least one", wait + "if ($s) { if (1 <= $s) { handle($s); } }", 0},
		{"nested reversed", wait + "if ($s) { if (0 < $s) { handle($s); } }", 0},
		{"nested conjunction", wait + "if ($s) { if ($x && $s >= 0) { handle($s); } }", 0},
		{"nested disjunction", wait + "if ($s) { if ($x || $s > 0) { handle($s); } }", 1},
		{"nested failure only", wait + "if ($s) { if ($s === -1) { handle($s); } }", 1},
		{"nested else", wait + "if ($s) { if ($s === -1) { echo 1; } else { handle($s); } }", 0},
		{"nested else positive", wait + "if ($s) { if ($s > 0) { echo 1; } else { handle($s); } }", 1},
		{"nested elseif", wait + "if ($s) { if ($x) { echo 1; } elseif ($s > 0) { handle($s); } }", 0},
		{"nested elseif earlier", wait + "if ($s) { if ($s < 0) { echo 1; } elseif ($x) { handle($s); } }", 0},
		{"nested else after elseif", wait + "if ($s) { if ($x) { echo 1; } elseif ($s == -1) { echo 2; } else { handle($s); } }", 0},
		{"nested second elseif", wait + "if ($s) { if ($x) { echo 1; } elseif ($s == -1) { echo 2; } elseif ($y) { handle($s); } }", 0},
		{"nested elseif plain", wait + "if ($s) { if ($x) { echo 1; } elseif ($y) { handle($s); } else { handle($s); } }", 1},
		{"ternary then", wait + "if ($s) { $r = $s > 0 ? handle($s) : null; }", 0},
		{"ternary else", wait + "if ($s) { $r = $s === -1 ? null : handle($s); }", 0},
		{"ternary wrong arm", wait + "if ($s) { $r = $s > 0 ? null : handle($s); }", 1},
		{"short ternary", wait + "if ($s) { $r = $s === -1 ?: handle($s); }", 0},
		{"ternary condition", wait + "if ($s) { $r = handle($s) ? 1 : 2; }", 1},
		{"and operand", wait + "if ($s) { $s > 0 && handle($s); }", 0},
		{"and keyword", wait + "if ($s) { $s > 0 and handle($s); }", 0},
		{"or operand", wait + "if ($s) { $s === -1 || handle($s); }", 0},
		{"or keyword", wait + "if ($s) { $s === -1 or handle($s); }", 0},
		{"or wrong side", wait + "if ($s) { $s > 0 || handle($s); }", 1},
		{"left operand", wait + "if ($s) { handle($s) && $s > 0; }", 1},
		{"other operator", wait + "if ($s) { $r = $x . handle($s); }", 1},
		{"condition operand", wait + "if ($s) { if ($s > 0 && handle($s)) { echo 1; } }", 0},
		{"condition call", wait + "if ($s) { if (handle($s)) { echo 1; } }", 1},
		{"alias guarded", wait + "$t = $s; if ($s) { if ($t > 0) { handle($s); } }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, source := range []string{
				"<?php function probe($set, $x, $y, $o, $n) { " + tc.body + " }",
				"<?php " + tc.body,
				"<?php namespace App; " + tc.body,
			} {
				f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe %s: %v", source, f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("%s: got %v, want %d", source, got, tc.findings)
				}
			}
		})
	}
}
