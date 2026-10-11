package parseurlfailuredereferenced

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewGuards(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const head = "<?php function f($u, $o) { $p = parse_url($u); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"falsy exit", "if (!$p) { return null; } return $p['host'];", 0},
		{"not array exit", "if (!is_array($p)) { return null; } return $p['host'];", 0},
		{"qualified not array exit", "if (!\\is_array($p)) { throw new Exception(); } return $p['host'];", 0},
		{"array branch", "if (is_array($p)) { return $p['host']; } return null;", 0},
		{"isset branch", "if (isset($p['host'])) { return $p['host']; } return null;", 0},
		{"isset exit", "if (!isset($p['scheme'], $p['host'])) { return; } echo $p['scheme'] . $p['host'];", 0},
		{"nested isset", "if (isset($p['a']['b'])) { echo $p['a']; }", 0},
		{"not empty key", "if (!empty($p['host'])) { echo $p['host']; }", 0},
		{"empty exit", "if (empty($p)) { return; } echo $p['host'];", 0},
		{"loose exit", "if ($p == false) { return; } echo $p['host'];", 0},
		{"loose yoda branch", "if (\\FALSE != $p) { echo $p['host']; }", 0},
		{"strict ternary", "return $p === false ? null : $p['host'];", 0},
		{"truthy ternary", "return $p ? $p['host'] : null;", 0},
		{"array ternary", "return is_array($p) ? $p['host'] : null;", 0},
		{"and chain", "return $p !== false && $p['host'] === 'x';", 0},
		{"keyword and chain", "return ($p and $p['host']);", 0},
		{"or chain", "return !$p || $p['host'] === 'x';", 0},
		{"keyword or chain", "return ($p === false or $p['host']);", 0},
		{"and condition", "if (is_array($p) && $p['host'] === 'x') { echo 1; }", 0},
		{"and branch", "if ($o && $p) { echo $p['host']; }", 0},
		{"or exit", "if (!is_array($p) || !isset($p['host'])) { return; } echo $p['host'];", 0},
		{"negated or branch", "if (!($p === false || $o)) { echo $p['host']; }", 0},
		{"and exit", "if (!$p && empty($p)) { return; } echo $p['host'];", 0},
		{"or branch", "if (is_array($p) || isset($p['x'])) { echo $p['host']; }", 0},
		{"else branch", "if (!$p) { echo 0; } else { echo $p['host']; }", 0},
		{"elseif branch", "if ($o) { echo 0; } elseif ($p) { echo $p['host']; }", 0},
		{"after failed elseif", "if ($o) { echo 0; } elseif (!$p) { echo 1; } else { echo $p['host']; }", 0},
		{"elseif after failed if", "if (!$p) { echo 0; } elseif ($o) { echo $p['host']; }", 0},
		{"exit with else", "if (!$p) { return; } else { echo 1; } echo $p['host'];", 0},
		{"earlier exit", "if (!$p) { return; } echo 1; echo $p['host'];", 0},
		{"outer exit", "if (!$p) { return; } if ($o) { echo $p['host']; }", 0},
		{"key exists", "if (array_key_exists('host', $p)) { echo $p['host']; }", 0},
		{"unbraced body", "if ($o) echo $p['host'];", 1},

		{"unrelated truthy", "if ($o) { echo $p['host']; }", 1},
		{"wrong branch", "if (!$p) { echo $p['host']; }", 1},
		{"wrong ternary arm", "return $p ? null : $p['host'];", 1},
		{"short ternary", "return $o ?: $p['host'];", 1},
		{"plain isset", "if (isset($p)) { echo $p['host']; }", 1},
		{"negated isset branch", "if (!isset($p['host'])) { echo $p['host']; }", 1},
		{"empty branch", "if (empty($p)) { echo $p['host']; }", 1},
		{"negated array branch", "if (!is_array($p)) { echo $p['host']; }", 1},
		{"other array", "if (is_array($o)) { echo $p['host']; }", 1},
		{"other key exists", "if (array_key_exists('host', $o)) { echo $p['host']; }", 1},
		{"other call", "if (count($p)) { echo $p['host']; }", 1},
		{"null comparison", "if ($p !== null) { echo $p['host']; }", 1},
		{"true comparison", "if ($p === true) { echo $p['host']; }", 1},
		{"identical branch", "if ($p === false) { echo $p['host']; }", 1},
		{"not identical exit", "if ($p !== false) { return; } echo $p['host'];", 1},
		{"relational", "if ($p > 0) { echo $p['host']; }", 1},
		{"negation operand", "if (-$p) { echo $p['host']; }", 1},
		{"or needs both", "if ($p || $o) { echo $p['host']; }", 1},
		{"and exit needs both", "if (!$p && $o) { return; } echo $p['host'];", 1},
		{"left operand", "return $p['host'] && $p;", 1},
		{"coalesce operand", "return $o ?? $p['host'];", 1},
		{"nonterminating exit", "if (!$p) { echo 0; } echo $p['host'];", 1},
		{"later exit", "echo $p['host']; if (!$p) { return; }", 1},
		{"condition read", "if ($p['host']) { echo 1; }", 1},
		{"elseif condition", "if ($o) { echo 0; } elseif ($p['host']) { echo 1; }", 1},
		{"unproven else", "if ($o) { echo 0; } elseif ($o) { echo 1; } else { echo $p['host']; }", 1},
		{"assert", "assert($p !== false); echo $p['host'];", 1},
		{"closure boundary", "if (!$p) { return; } return function ($p) { return $p; };", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(head+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}

func TestReviewGuardValueIdentity(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"guard before producer", "<?php function f($u, $p) { if (!$p) { return; } $p = parse_url($u); echo $p['host']; }", 1},
		{"guard of earlier result", "<?php function f($u, $v) { $p = parse_url($u); if (!$p) { return; } $p = parse_url($v); echo $p['host']; }", 1},
		{"direct call", "<?php function f($u) { if ($u) { echo parse_url($u)['host']; } }", 1},
		{"file scope", "<?php $p = parse_url($u); if (!$p) { exit(1); } echo $p['host'];", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
