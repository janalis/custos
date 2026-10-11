package opensslsigningfailureignored

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
	const sign = "openssl_sign($d, $s, $k); "
	const publish = "echo base64_encode($s);"
	fn := func(body string) string { return "<?php function f($d, $k, $p) { " + body + " }" }
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		// Tests that prove success on a variable known to be unset or empty.
		{"empty guard", fn(sign + "if (empty($s)) { return; } " + publish), 0},
		{"negated isset guard", fn(sign + "if (!isset($s)) { return; } " + publish), 0},
		{"isset list", fn(sign + "if (!isset($d, $s)) { return; } " + publish), 0},
		{"negated guard", fn(sign + "if (!$s) { throw new Exception(); } " + publish), 0},
		{"either failure", fn(sign + "if ($d === '' || empty($s)) { return; } " + publish), 0},
		{"guard then nested use", fn(sign + "if (empty($s)) { return; } if ($d) { " + publish + " }"), 0},
		{"guard after other statement", fn(sign + "$n = 1; if (empty($s)) { return; } " + publish), 0},
		{"truthy branch", fn(sign + "if ($s) { " + publish + " }"), 0},
		{"parenthesised truthy branch", fn(sign + "if (($s)) { " + publish + " }"), 0},
		{"truthy conjunction", fn(sign + "if ($s && $d) { " + publish + " }"), 0},
		{"truthy ternary", fn(sign + "echo $s ? base64_encode($s) : '';"), 0},
		{"else of empty", fn(sign + "echo empty($s) ? '' : bin2hex($s);"), 0},
		{"short circuit", fn(sign + "isset($s) && file_put_contents('sig', $s);"), 0},
		{"null initialised", fn("$s = null; " + sign + "if (!isset($s)) { return; } " + publish), 0},
		{"first statement of closure", "<?php $f = function ($d, $k) { " + sign + "if (empty($s)) { return; } " + publish + " };", 0},
		{"other variables", fn(sign + "if (empty($s)) { return; } $g = fn ($s) => $s; " + publish), 0},
		{"capture after guard", fn(sign + "if (empty($s)) { return; } $g = function () use ($s, $d) { return $s; }; " + publish), 0},
		// openssl_verify() is accepted whatever the variable held.
		{"verify in loop", fn("while ($d) { " + sign + "if (1 !== openssl_verify($d, $s, $p)) { break; } " + publish + " }"), 0},
		{"verify branch", "<?php " + sign + "if (openssl_verify($d, ($s), $p) === 1) { " + publish + " }", 0},
		{"verify of parameter", "<?php function g($d, $s, $k, $p) { " + sign + "if (openssl_verify($d, $s, $p) !== 1) { return; } " + publish + " }", 0},
		// Stale or unknown content: the emptiness test proves nothing.
		{"file scope", "<?php " + sign + "if (empty($s)) { return; } " + publish, 1},
		{"file scope ternary", "<?php " + sign + "echo isset($s) ? base64_encode($s) : '';", 1},
		{"earlier assignment", fn("$s = 'old'; " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"earlier compound assignment", fn("$s .= ''; " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"earlier reference", fn("$s = &$d; " + sign + "if (empty($s)) { return; } " + publish), 0}, // nothing writes either name after the test
		{"earlier call", fn("prepare(); " + sign + "if (empty($s)) { return; } " + publish), 0},
		{"earlier echo", fn("echo $s; " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"dynamic variable", fn("$$d = 'old'; " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"label", fn("again: " + sign + "if (empty($s)) { goto again; } " + publish), 1},
		{"include", fn("include 'state.php'; " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"eval", fn("eval($d); " + sign + "if (empty($s)) { return; } " + publish), 1},
		{"do loop", fn("do { if ($d) { " + sign + "if (empty($s)) { continue; } " + publish + " } } while ($d);"), 1},
		{"for loop", fn("for (;;) { " + sign + "if (empty($s)) { break; } " + publish + " }"), 1},
		{"test before call", fn("if (openssl_verify($d, $s, $p) === 1) { " + sign + publish + " }"), 1},
		// Tests that do not establish success.
		{"other variable truthy", fn(sign + "if ($d) { " + publish + " }"), 1},
		{"isset of other", fn(sign + "if (isset($d)) { " + publish + " }"), 1},
		{"empty of other", fn(sign + "if (empty($d)) { return; } " + publish), 1},
		{"negative number", fn(sign + "if (-$d) { return; } " + publish), 1},
		{"call condition", fn(sign + "if (check($d)) { " + publish + " }"), 1},
		{"length condition", fn(sign + "if (strlen($s) > 0) { " + publish + " }"), 1},
		{"either success", fn(sign + "if ($d || $s) { " + publish + " }"), 1},
		{"both failures", fn(sign + "if ($d && empty($s)) { return; } " + publish), 1},
		{"non terminating guard", fn(sign + "if (empty($s)) { log(); } " + publish), 1},
		{"guard with else", fn(sign + "if (empty($s)) { return; } else { log(); } " + publish), 1},
		{"else after elseif", fn(sign + "if (empty($s)) { return; } elseif ($d) { log(); } else { " + publish + " }"), 1},
		{"elseif body", fn(sign + "if ($d) { log(); } elseif ($k) { " + publish + " }"), 1},
		{"short ternary", fn(sign + "$t = $s ?: 'none'; " + publish), 1},
		{"unguarded right operand", fn(sign + "$d && file_put_contents('sig', $s);"), 1},
		{"cast before publishing", fn(sign + "echo base64_encode((string) $s);"), 0},
		{"concatenation", fn(sign + "$t = $s . ''; " + publish), 1},
		{"verify loosely compared", fn(sign + "if (!openssl_verify($d, $s, $p)) { return; } " + publish), 1},
		{"verify equals zero", fn(sign + "if (openssl_verify($d, $s, $p) !== 0) { return; } " + publish), 1},
		{"verify against name", fn(sign + "if (openssl_verify($d, $s, $p) !== $d) { return; } " + publish), 1},
		{"verify of other value", fn(sign + "if (openssl_verify($d, $d, $p) !== 1) { return; } " + publish), 1},
		{"verify without arguments", fn(sign + "if (openssl_verify() !== 1) { return; } " + publish), 1},
		{"other call compared", fn(sign + "if (check($d) !== 1) { return; } " + publish), 1},
		{"variable compared", fn(sign + "if ($d !== 1) { return; } " + publish), 1},
		{"loose comparison", fn(sign + "if (openssl_verify($d, $s, $p) != 1) { return; } " + publish), 1},
		// Shapes the guard analysis does not decide keep the base behaviour.
		{"property signature", fn("openssl_sign($d, $this->s, $k); echo base64_encode($this->s);"), 0},
		{"dynamic signature", fn("openssl_sign($d, $$d, $k); echo base64_encode($$d);"), 1},
		{"captured by closure", fn("$g = function () use (&$s) { $s = 'old'; }; $g(); " + sign + "if (empty($s)) { return; } " + publish), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
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
