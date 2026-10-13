package realpathfailureusedaspath

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		php        phpversion.Version
		findings   int
	}{
		{"negated exit", "$r = realpath($d); if (!$r) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"guard not adjacent", "$r = realpath($d); if (!$r) { return; } log_it(); is_file($r . '/x');", phpversion.PHP85, 0},
		{"truthy branch", "$r = realpath($d); if ($r) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"is_string branch", "$r = realpath($d); if (is_string($r)) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"not is_string exit", "$r = realpath($d); if (!is_string($r)) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"is_string other", "$r = realpath($d); if (is_string($o)) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"other call", "$r = realpath($d); if (is_dir($r)) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"loose exit", "$r = realpath($d); if ($r == false) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"loose yoda branch", "$r = realpath($d); if (false != $r) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"compared to true", "$r = realpath($d); if ($r !== true) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"compared to other", "$r = realpath($d); if ($o === false) { return; } is_file($r . '/x');", phpversion.PHP85, 1},
		{"property truthiness", "$r = realpath($d); if ($o->ready) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"relational", "$r = realpath($d); if ($r > false) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"empty exit", "$r = realpath($d); if (empty($r)) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"empty branch", "$r = realpath($d); if (empty($r)) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"ternary else", "$r = realpath($d); echo $r === false ? 'no' : is_file($r . '/x');", phpversion.PHP85, 0},
		{"ternary then", "$r = realpath($d); echo $r ? is_file($r . '/x') : 'no';", phpversion.PHP85, 0},
		{"ternary wrong arm", "$r = realpath($d); echo $r ? 'no' : is_file($r . '/x');", phpversion.PHP85, 1},
		{"ternary condition", "$r = realpath($d); echo is_file($r . '/x') ? $r : 'no';", phpversion.PHP85, 1},
		{"short ternary", "$r = realpath($d); echo $o ?: is_file($r . '/x');", phpversion.PHP85, 1},
		{"disjunct exit", "$r = realpath($d); if (!$r || !is_dir($r)) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"second disjunct exit", "$r = realpath($d); if ($o or !$r) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"conjunct exit", "$r = realpath($d); if (!$r && $o) { return; } is_file($r . '/x');", phpversion.PHP85, 1},
		{"conjunct branch", "$r = realpath($d); if ($r && is_dir($r)) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"second conjunct branch", "$r = realpath($d); if ($o and $r) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"disjunct branch", "$r = realpath($d); if ($r || $o) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"xor branch", "$r = realpath($d); if ($r xor $o) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"inline conjunction", "$r = realpath($d); echo $r && is_file($r . '/x');", phpversion.PHP85, 0},
		{"inline disjunction", "$r = realpath($d); echo !$r || is_file($r . '/x');", phpversion.PHP85, 0},
		{"inline wrong side", "$r = realpath($d); echo $r || is_file($r . '/x');", phpversion.PHP85, 1},
		{"inline left operand", "$r = realpath($d); echo is_file($r . '/x') && $r;", phpversion.PHP85, 1},
		{"inline coalesce", "$r = realpath($d); echo $o ?? is_file($r . '/x');", phpversion.PHP85, 1},
		{"throw expression", "$r = realpath($d); $r !== false || throw new \\RuntimeException(); is_file($r . '/x');", phpversion.PHP85, 0},
		{"throw expression before 8.0", "$r = realpath($d); $r !== false || throw new \\RuntimeException(); is_file($r . '/x');", phpversion.PHP74, 1},
		{"and exit", "$r = realpath($d); $r === false && exit(1); is_file($r . '/x');", phpversion.PHP85, 0},
		{"or die", "$r = realpath($d) or die('x'); is_file($r . '/x');", phpversion.PHP74, 0},
		{"abort on success", "$r = realpath($d); $r !== false && exit(1); is_file($r . '/x');", phpversion.PHP85, 1},
		{"abort on truthy", "$r = realpath($d); !$r || exit(1); is_file($r . '/x');", phpversion.PHP85, 1},
		{"xor abort", "$r = realpath($d); $r xor exit(1); is_file($r . '/x');", phpversion.PHP85, 1},
		{"or continues", "$r = realpath($d) or log_it(); is_file($r . '/x');", phpversion.PHP85, 1},
		{"unrelated statement", "$r = realpath($d); $o = 1 + 2; is_file($r . '/x');", phpversion.PHP85, 1},
		{"assignment in condition", "if (!($r = realpath($d))) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"other assignment in condition", "$r = realpath($d); if (!($o = realpath($e))) { return; } is_file($r . '/x');", phpversion.PHP85, 1},
		{"non-terminating guard", "$r = realpath($d); if (!$r) { log_it(); } is_file($r . '/x');", phpversion.PHP85, 1},
		{"guard with else", "$r = realpath($d); if (!$r) { return; } else { log_it(); } is_file($r . '/x');", phpversion.PHP85, 0},
		{"guard with elseif", "$r = realpath($d); if (!$r) { return; } elseif ($o) { log_it(); } is_file($r . '/x');", phpversion.PHP85, 0},
		{"else branch", "$r = realpath($d); if (!$r) { log_it(); } else { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"else of truthy", "$r = realpath($d); if ($r) { log_it(); } else { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"elseif body", "$r = realpath($d); if ($o) { log_it(); } elseif ($r) { is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"elseif condition", "$r = realpath($d); if ($r === false) { log_it(); } elseif (is_file($r . '/x')) { log_it(); }", phpversion.PHP85, 0},
		{"after failed elseif", "$r = realpath($d); if ($o) { log_it(); } elseif (!$r) { log_it(); } elseif ($p) { is_file($r . '/x'); } else { log_it(); }", phpversion.PHP85, 0},
		{"before later elseif", "$r = realpath($d); if ($o) { is_file($r . '/x'); } elseif (!$r) { log_it(); }", phpversion.PHP85, 1},
		{"unguarded elseif", "$r = realpath($d); if ($o) { log_it(); } elseif ($p) { is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"if condition", "$r = realpath($d); if (is_file($r . '/x')) { log_it(); }", phpversion.PHP85, 1},
		{"while body", "$r = realpath($d); while ($r !== false) { return is_file($r . '/x'); }", phpversion.PHP85, 0},
		{"while other", "$r = realpath($d); while ($o) { return is_file($r . '/x'); }", phpversion.PHP85, 1},
		{"while condition", "$r = realpath($d); while (is_file($r . '/x')) { log_it(); }", phpversion.PHP85, 1},
		{"assert is not a guard", "$r = realpath($d); assert($r !== false); is_file($r . '/x');", phpversion.PHP85, 1},
		{"stale guard", "$r = realpath($d); if (!$r) { return; } $r = realpath($e); is_file($r . '/x');", phpversion.PHP85, 1},
		{"join before guard", "$r = realpath($d); $p = $r . '/x'; if (!$r) { return; } is_file($p);", phpversion.PHP85, 0},
		{"join before unrelated guard", "$r = realpath($d); $p = $r . '/x'; if (!$o) { return; } is_file($p);", phpversion.PHP85, 1},
		// The flow facts are read where the result is joined; a check written
		// after the join is found by the rule's own walk from the consuming call.
		{"join before branch", "$r = realpath($d); $p = $r . '/x'; if ($r) { is_file($p); }", phpversion.PHP85, 0},
		{"join before unrelated branch", "$r = realpath($d); $p = $r . '/x'; if ($o) { is_file($p); }", phpversion.PHP85, 1},
		{"join before else", "$r = realpath($d); $p = $r . '/x'; if (!$r) { log_it(); } else { is_file($p); }", phpversion.PHP85, 0},
		{"join before else of truthy", "$r = realpath($d); $p = $r . '/x'; if ($r) { log_it(); } else { is_file($p); }", phpversion.PHP85, 1},
		{"join before failed elseif", "$r = realpath($d); $p = $r . '/x'; if ($o) { log_it(); } elseif (!$r) { log_it(); } elseif ($q) { is_file($p); }", phpversion.PHP85, 0},
		{"join before later elseif", "$r = realpath($d); $p = $r . '/x'; if ($o) { is_file($p); } elseif (!$r) { log_it(); }", phpversion.PHP85, 1},
		{"join before elseif body", "$r = realpath($d); $p = $r . '/x'; if ($o) { log_it(); } elseif ($r) { is_file($p); }", phpversion.PHP85, 0},
		{"join before while body", "$r = realpath($d); $p = $r . '/x'; while ($r !== false) { return is_file($p); }", phpversion.PHP85, 0},
		{"join before ternary then", "$r = realpath($d); $p = $r . '/x'; echo $r ? is_file($p) : 'no';", phpversion.PHP85, 0},
		{"join before ternary else", "$r = realpath($d); $p = $r . '/x'; echo $r === false ? 'no' : is_file($p);", phpversion.PHP85, 0},
		{"join before ternary wrong arm", "$r = realpath($d); $p = $r . '/x'; echo $r ? 'no' : is_file($p);", phpversion.PHP85, 1},
		{"join before inline conjunction", "$r = realpath($d); $p = $r . '/x'; echo $r && is_file($p);", phpversion.PHP85, 0},
		{"join before inline disjunction", "$r = realpath($d); $p = $r . '/x'; echo !$r || is_file($p);", phpversion.PHP85, 0},
		{"join before inline wrong side", "$r = realpath($d); $p = $r . '/x'; echo $r || is_file($p);", phpversion.PHP85, 1},
		{"join before or exit", "$r = realpath($d); $p = $r . '/x'; $r !== false || exit(1); is_file($p);", phpversion.PHP85, 0},
		{"join before and exit", "$r = realpath($d); $p = $r . '/x'; $r === false && exit(1); is_file($p);", phpversion.PHP85, 0},
		{"join before abort on success", "$r = realpath($d); $p = $r . '/x'; $r !== false && exit(1); is_file($p);", phpversion.PHP85, 1},
		// The path comes from a project call or a property: the result is not
		// completely known, the facts of its paths still hold.
		{"project call path", "$r = realpath(load($d)); is_file($r . '/x');", phpversion.PHP85, 1},
		{"project call path guarded", "$r = realpath(load($d)); if (!$r) { return; } is_file($r . '/x');", phpversion.PHP85, 0},
		{"property path", "$r = realpath($o->dir); is_file($r . '/x');", phpversion.PHP85, 1},
		{"property path guarded", "$r = realpath($o->dir); $r !== false or exit; is_file($r . '/x');", phpversion.PHP85, 0},
		// A scope with `goto` is not followed by the flow: the walk alone decides.
		{"unfollowed scope", "$r = realpath($d); is_file($r . '/x'); goto done; done: ;", phpversion.PHP85, 1},
		{"unfollowed scope guarded", "$r = realpath($d); if (!$r) { return; } is_file($r . '/x'); goto done; done: ;", phpversion.PHP85, 0},
		{"guard in closure scope", "$r = realpath($d); if (!$r) { return; } $f = function () use ($d) { $r = realpath($d); is_file($r . '/x'); };", phpversion.PHP85, 1},
		{"direct call", "if ($o) { is_file(realpath($d) . '/x'); }", phpversion.PHP85, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, Only: []string{"RealpathFailureUsedAsPath"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.ParseBest("review.php", []byte("<?php\n"+tc.body+"\n"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
