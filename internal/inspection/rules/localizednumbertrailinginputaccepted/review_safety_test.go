package localizednumbertrailinginputaccepted

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
		decimal = `$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); `
		parse   = `$s = "37tail"; $n = $f->parse($s, NumberFormatter::TYPE_DOUBLE, $end); `
	)
	for _, tc := range []struct {
		name, source string
		// span is the reported source text; empty means no finding.
		span string
	}{
		{"plain decimal", decimal + parse + `echo $n;`, "echo $n"},
		{"close tag terminator", decimal + parse + `echo $n ?>`, "echo $n"},
		{"parenthesized value", decimal + parse + `echo ($n) ;`, "echo ($n)"},
		{"integer style", `$f = new NumberFormatter("en_US", 1); ` + parse + `echo $n;`, "echo $n"},
		{"named arguments", `$f = new NumberFormatter(style: NumberFormatter::DECIMAL, locale: "en_US"); ` + parse + `echo $n;`, "echo $n"},
		{"inline construction", `$n = (new NumberFormatter("en_US", NumberFormatter::DECIMAL))->parse("37tail"); echo $n;`, "echo $n"},
		{"passive calls", decimal + `$f->getErrorCode(); $f->format(1); ` + parse + `echo $n;`, "echo $n"},
		{"configured afterwards", decimal + parse + `$f->setPattern("#0'tail'"); echo $n;`, "echo $n"},

		{"suffix pattern", `$f = new NumberFormatter("en_US", NumberFormatter::PATTERN_DECIMAL, "#0'tail'"); ` + parse + `echo $n;`, ""},
		{"decimal with pattern", `$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL, "#0'tail'"); ` + parse + `echo $n;`, ""},
		{"other style", `$f = new NumberFormatter("en_US", NumberFormatter::SPELLOUT); ` + parse + `echo $n;`, ""},
		{"unknown style", `function f($style) { $f = new NumberFormatter("en_US", $style); ` + parse + `echo $n; }`, ""},
		{"missing arguments", `$f = new NumberFormatter; ` + parse + `echo $n;`, ""},
		{"missing style", `$f = new NumberFormatter(locale: "en_US", pattern: "#"); ` + parse + `echo $n;`, ""},
		{"suffix attribute", decimal + `$f->setTextAttribute(NumberFormatter::POSITIVE_SUFFIX, "tail"); ` + parse + `echo $n;`, ""},
		{"pattern setter", decimal + `$f->setPattern("#0'tail'"); ` + parse + `echo $n;`, ""},
		{"lenient attribute", decimal + `$f->setAttribute(NumberFormatter::LENIENT_PARSE, 1); ` + parse + `echo $n;`, ""},
		{"dynamic method", decimal + `$f->$m(); ` + parse + `echo $n;`, ""},
		{"procedural setter", decimal + `numfmt_set_pattern($f, "#0'tail'"); ` + parse + `echo $n;`, ""},
		{"alias", decimal + `$g = $f; $g->setPattern("#0'tail'"); ` + parse + `echo $n;`, ""},
		{"captured", decimal + `$c = function () use ($f) { $f->setPattern("#0'tail'"); }; $c(); ` + parse + `echo $n;`, ""},
		{"configured later in loop", decimal + `foreach ([1, 2] as $i) { ` + parse + `echo $n; $f->setPattern("#0'tail'"); }`, ""},
		{"configured later in while", decimal + `while ($i--) { ` + parse + `echo $n; $f->setPattern("#0'tail'"); }`, ""},
		{"configured later in for", decimal + `for (;;) { ` + parse + `echo $n; $f->setPattern("#0'tail'"); }`, ""},
		{"configured later in do", decimal + `do { ` + parse + `echo $n; $f->setPattern("#0'tail'"); } while ($i--);`, ""},
		{"inline pattern", `$n = (new NumberFormatter("en_US", NumberFormatter::PATTERN_DECIMAL, "#0'tail'"))->parse("37tail"); echo $n;`, ""},

		{"exit guard", decimal + parse + `if ($end !== strlen($s)) { exit("bad"); } echo $n;`, ""},
		{"throw guard", decimal + parse + `if (strlen($s) !== $end) { throw new RuntimeException("bad"); } echo $n;`, ""},
		{"return guard", `function f() { ` + decimal + parse + `if (!($end === strlen($s))) return; echo $n; }`, ""},
		{"combined exit guard", decimal + parse + `if ($n === false || $end !== strlen($s)) { exit; } echo $n;`, ""},
		{"outer exit guard", decimal + parse + `if ($end !== strlen($s)) { exit; } if ($n !== false) { echo $n; }`, ""},
		{"word operators", decimal + parse + `if ($n !== false and $end === strlen($s)) { echo $n; }`, ""},
		{"word exit guard", decimal + parse + `if ($n === false or $end !== strlen($s)) { exit; } echo $n;`, ""},
		{"else branch", decimal + parse + `if ($end !== strlen($s)) { error_log("bad"); } else { echo $n; }`, ""},
		{"elseif branch", decimal + parse + `if ($n === false) { exit; } elseif ($end === strlen($s)) { echo $n; }`, ""},
		{"guard with else", decimal + parse + `if ($end !== strlen($s)) { exit; } else { error_log("ok"); } echo $n;`, ""},
		{"guard with elseif", decimal + parse + `if ($end !== strlen($s)) { exit; } elseif ($n) { error_log("ok"); } echo $n;`, ""},
		{"length constant", decimal + parse + `if ($end !== 6) { exit; } echo $n;`, ""},
		{"short offset exits", decimal + parse + `if ($end < strlen($s)) { exit; } echo $n;`, ""},
		{"saved comparison", decimal + parse + `$whole = $end === strlen($s); if ($whole) { echo $n; }`, ""},
		{"jump label after guard", decimal + parse + `if ($end !== strlen($s)) { exit; } again: echo $n; goto again;`, "echo $n"},
		{"reversed operands", decimal + parse + `if (strlen($s) === $end) { echo $n; }`, ""},

		{"wrong branch", decimal + parse + `if ($end === strlen($s)) { error_log("ok"); } else { echo $n; }`, "echo $n"},
		{"wrong elseif branch", decimal + parse + `if ($end === strlen($s)) { exit; } elseif ($n !== false) { echo $n; }`, "echo $n"},
		{"mismatch branch", decimal + parse + `if ($end !== strlen($s)) { echo $n; }`, "echo $n"},
		{"alternative condition", decimal + parse + `if ($n !== false || $end === strlen($s)) { echo $n; }`, "echo $n"},
		{"optional exit guard", decimal + parse + `if ($n === false && $end !== strlen($s)) { exit; } echo $n;`, "echo $n"},
		{"guard continues", decimal + parse + `if ($end !== strlen($s)) { error_log("bad"); } echo $n;`, "echo $n"},
		{"guard before parse", decimal + `$s = "37tail"; $end = 0; if ($end !== strlen($s)) { exit; } $n = $f->parse($s, NumberFormatter::TYPE_DOUBLE, $end); echo $n;`, "echo $n"},
		{"other input", decimal + parse + `if ($end !== strlen($t)) { exit; } echo $n;`, "echo $n"},
		{"other offset", decimal + parse + `if ($pos !== strlen($s)) { exit; } echo $n;`, "echo $n"},
		{"other length", decimal + parse + `if ($end !== mb_strlen($s)) { exit; } echo $n;`, "echo $n"},
		{"unrelated conditions", decimal + parse + `if (-$end) { exit; } if ($end > 1) { exit; } echo $n;`, "echo $n"},
		{"missing offset", decimal + `$s = "37tail"; $n = $f->parse($s); if ($end !== strlen($s)) { exit; } echo $n;`, "echo $n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("<?php " + tc.source)
			f := syntax.ParseBest("review.php", src, syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if tc.span == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected findings: %v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("findings: got %v, want 1", got)
			}
			if text := string(src[got[0].Span.Start:got[0].Span.End]); text != tc.span {
				t.Fatalf("range: got %q, want %q", text, tc.span)
			}
			if len(got[0].Fixes) != 0 {
				t.Fatal("unexpected fix")
			}
		})
	}
}
