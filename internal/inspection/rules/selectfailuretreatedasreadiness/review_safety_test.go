package selectfailuretreatedasreadiness

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
	const sel = "<?php function f($a, $b, $c, $o) { $r = stream_select($a, $b, $c, 1); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"unguarded", "if ($r !== 0) { echo 1; } }", 1},
		{"terminating guard", "if ($r === false) { throw new RuntimeException(); } if ($r !== 0) { echo 1; } }", 0},
		{"enclosing guard", "if ($r !== false) { if ($r !== 0) { echo 1; } } }", 0},
		{"else of failure test", "if (false === $r) { echo 0; } else { if ($r !== 0) { echo 1; } } }", 0},
		{"and guard", "if ($r !== false && $r !== 0) { echo 1; } }", 0},
		{"keyword and guard", "if ($r !== false and $r !== 0) { echo 1; } }", 0},
		{"or guard", "if (false === $r || $r !== 0 && count($a)) { echo 1; } }", 0},
		{"keyword or guard", "if ($r === false or $r !== 0) { echo 1; } }", 0},
		{"nested and guard", "if (($o && $r !== \\false) && ($o || $r !== 0)) { echo 1; } }", 0},
		{"nested or guard", "if (($o || $r === false) || $r !== 0) { echo 1; } }", 0},
		{"elseif guard", "if ($r === false) { return; } elseif ($r !== 0) { echo 1; } }", 0},
		{"later elseif guard", "if ($o) { return; } elseif ($r === false) { return; } elseif ($o > 1) { echo 2; } elseif ($r !== 0) { echo 1; } }", 0},
		{"elseif body guard", "if ($r === false) { return; } elseif ($o) { if ($r !== 0) { echo 1; } } }", 0},
		{"guard on left operand only", "if ($r !== 0 && $r !== false) { echo 1; } }", 1},
		{"or with inequality", "if ($r !== false || $r !== 0) { echo 1; } }", 1},
		{"and with identity", "if ($r === false && $r !== 0) { echo 1; } }", 1},
		{"disjunctive guard", "if (($o || $r !== false) && $r !== 0) { echo 1; } }", 1},
		{"conjunctive failure test", "if (($o && $r === false) || $r !== 0) { echo 1; } }", 1},
		{"other variable", "if ($o !== false && $r !== 0) { echo 1; } }", 1},
		{"null guard", "if ($r !== null && $r !== 0) { echo 1; } }", 1},
		{"loose guard operand", "if ($o && $r !== 0) { echo 1; } }", 1},
		{"unrelated operator", "if ($o + ($r !== 0)) { echo 1; } }", 1},
		{"elseif without failure test", "if ($o) { return; } elseif ($o > 1) { echo 2; } elseif ($r !== 0) { echo 1; } }", 1},
		{"failure test in later arm", "if ($o) { return; } elseif ($r !== 0) { echo 1; } elseif ($r === false) { return; } }", 1},
		{"guard before rebinding", "if ($r === false) { return; } $r = stream_select($a, $b, $c, 1); if ($r !== 0) { echo 1; } }", 1},
		{"guard on earlier result", "$q = $r; $r = stream_select($a, $b, $c, 1); if ($q !== false && $r !== 0) { echo 1; } }", 1},
		{"direct calls", "if (stream_select($a, $b, $c, 1) !== false && stream_select($a, $b, $c, 1) !== 0) { echo 1; } }", 1},
		{"direct call after guard", "if (stream_select($a, $b, $c, 1) === false) { return; } if (stream_select($a, $b, $c, 1) !== 0) { echo 1; } }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(sel+tc.body), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
