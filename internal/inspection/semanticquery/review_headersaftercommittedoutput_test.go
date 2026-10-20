package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewHeadersProbe dispatches on function calls only, as the rule package does.
type reviewHeadersProbe struct{ nativeCheckProbe }

func (reviewHeadersProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func TestReviewHeadersAfterCommittedOutput(t *testing.T) {
	const sent = "echo 'x'; flush();"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"header", sent + "header('X: 1');", 1},
		{"setcookie", sent + "setcookie('a', 'b');", 1},
		{"status setter", sent + "http_response_code(404);", 1},
		{"named status setter", sent + "http_response_code(response_code: 404);", 1},
		{"dynamic status setter", "function f($c){" + sent + "http_response_code($c);}", 1},
		{"status getter", sent + "$c = http_response_code();", 0},
		{"zero status reads", sent + "$c = http_response_code(0);", 0},
		{"null status reads", sent + "$c = http_response_code(null);", 0},
		{"unpacked status is unknown", "function f($a){" + sent + "http_response_code(...$a);}", 0},
		{"negated guard", sent + "if (!headers_sent()) { header('X: 1'); }", 0},
		{"guard with arguments", sent + "if (!headers_sent($file, $line)) { header('X: 1'); }", 0},
		{"else of positive guard", sent + "if (headers_sent()) { $x = 1; } else { header('X: 1'); }", 0},
		{"positive guard body", sent + "if (headers_sent()) { header('X: 1'); }", 1},
		{"else of negated guard", sent + "if (!headers_sent()) { $x = 1; } else { header('X: 1'); }", 1},
		{"early return", "function f(){" + sent + "if (headers_sent()) { return; } header('X: 1');}", 0},
		{"early return wrong polarity", "function f(){" + sent + "if (!headers_sent()) { return; } header('X: 1');}", 1},
		{"non-terminating guard", "function f(){" + sent + "if (headers_sent()) { $x = 1; } header('X: 1');}", 1},
		{"or", sent + "headers_sent() or header('X: 1');", 0},
		{"boolean or", sent + "headers_sent() || header('X: 1');", 0},
		{"and", sent + "!headers_sent() and header('X: 1');", 0},
		{"boolean and", sent + "!headers_sent() && header('X: 1');", 0},
		{"and wrong polarity", sent + "headers_sent() && header('X: 1');", 1},
		{"or wrong polarity", sent + "!headers_sent() || header('X: 1');", 1},
		{"left operand is not guarded", sent + "header('X: 1') || headers_sent();", 1},
		{"unrelated operator", sent + "$x = headers_sent() . header('X: 1');", 1},
		{"ternary then", sent + "$x = !headers_sent() ? header('X: 1') : null;", 0},
		{"ternary else", sent + "$x = headers_sent() ? null : header('X: 1');", 0},
		{"ternary wrong branch", sent + "$x = headers_sent() ? header('X: 1') : null;", 1},
		{"ternary condition", sent + "$x = header('X: 1') ? 1 : 2;", 1},
		{"elseif", "function f($a){" + sent + "if ($a) { $x = 1; } elseif (!headers_sent()) { header('X: 1'); }}", 0},
		{"elseif wrong polarity", "function f($a){" + sent + "if ($a) { $x = 1; } elseif (headers_sent()) { header('X: 1'); }}", 1},
		{"elseif condition", "function f($a){" + sent + "if ($a) { $x = 1; } elseif (header('X: 1')) { $x = 1; }}", 1},
		{"conjunction guard", "function f($a){" + sent + "if ($a && !headers_sent()) { header('X: 1'); }}", 0},
		{"conjunction guard left", "function f($a){" + sent + "if (!headers_sent() and $a) { header('X: 1'); }}", 0},
		{"disjunction guard proves nothing", "function f($a){" + sent + "if ($a || !headers_sent()) { header('X: 1'); }}", 1},
		{"disjunction early return", "function f($a){" + sent + "if ($a || headers_sent()) { return; } header('X: 1');}", 0},
		{"conjunction early return proves nothing", "function f($a){" + sent + "if ($a && headers_sent()) { return; } header('X: 1');}", 1},
		{"identical false", sent + "if (headers_sent() === false) { header('X: 1'); }", 0},
		{"yoda not identical true", sent + "if (true !== headers_sent()) { header('X: 1'); }", 0},
		{"equal true", sent + "if (headers_sent() == true) { header('X: 1'); }", 1},
		{"not equal false", sent + "if (headers_sent() != false) { header('X: 1'); }", 1},
		{"comparison with non-boolean", "function f($a){" + sent + "if (headers_sent() === $a) { header('X: 1'); }}", 1},
		{"other comparison", sent + "if (headers_sent() < 1) { header('X: 1'); }", 1},
		{"other unary", sent + "if (-headers_sent()) { $x = 1; } else { header('X: 1'); }", 1},
		{"other condition", "function f($a){" + sent + "if (!$a) { header('X: 1'); }}", 1},
		{"other function", sent + "if (!headers_list()) { header('X: 1'); }", 1},
		{"commit after guard", "if (!headers_sent()) { echo 'x'; flush(); header('X: 1'); }", 1},
		{"commit after short-circuit guard", "echo 'x'; headers_sent() or (flush() xor header('X: 1'));", 1},
		{"guard does not cross scopes", sent + "if (!headers_sent()) { $f = function () { echo 'y'; flush(); header('X: 1'); }; }", 1},
		{"saved test", "function f(){" + sent + "$sent = headers_sent(); if (!$sent) { header('X: 1'); }}", 0},
		{"early return with else", "function f(){" + sent + "if (headers_sent()) { return; } else { $x = 1; } header('X: 1');}", 0},
		{"while guard", "function f(){" + sent + "while (!headers_sent()) { header('X: 1'); break; }}", 0},
		{"match arm", "function f(){" + sent + "return match (true) { !headers_sent() => header('X: 1'), default => null };}", 0},
		{"commit between saved test and call", "function f(){$sent = headers_sent(); " + sent + "if (!$sent) { header('X: 1'); }}", 1},
		{"buffered output", "ob_start(); " + sent + "header('X: 1');", 0},
		{"no output", "flush(); header('X: 1');", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewHeadersProbe{nativeCheckProbe{"HeadersAfterCommittedOutput", CheckHeadersAfterCommittedOutput}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"HeadersAfterCommittedOutput"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
