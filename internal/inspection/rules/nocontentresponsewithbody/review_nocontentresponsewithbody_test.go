package nocontentresponsewithbody

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewNoContentResponseWithBody(t *testing.T) {
	const set = "http_response_code(204);"
	const body = "echo 'body';"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"plain body", set + body, 1},
		{"inside function", "function f(){" + set + body + "}", 1},
		{"empty body", set + "echo '';", 0},
		{"other status", "http_response_code(200);" + body, 0},
		{"dynamic status", "function f($c){http_response_code($c);" + body + "}", 0},

		{"status line", set + "header('HTTP/1.1 200 OK');" + body, 0},
		{"lower-case status line", set + "header('http/1.0 404 Not Found');" + body, 0},
		{"status line repeating 204", set + "header('HTTP/1.1 204 No Content');" + body, 1},
		{"status line without code", set + "header('HTTP/1.1');" + body, 0},
		{"status line with dynamic tail", "function f($r){" + set + "header('HTTP/1.1 204 ' . $r);" + body + "}", 0},
		{"location field", set + "header('Location: /next');" + body, 0},
		{"spaced location field", set + "header('location : /next');" + body, 0},
		{"status field", set + "header('Status: 200');" + body, 0},
		{"unrelated field", set + "header('Content-Type: text/plain');" + body, 1},
		{"line without field name", set + "header('X');" + body, 1},
		{"dynamic line", "function f($h){" + set + "header($h);" + body + "}", 0},
		{"line through variable", set + "$h = 'HTTP/1.1 200 OK'; header($h);" + body, 0},
		{"concatenated value", "function f($t){" + set + "header('Content-Type: ' . $t . '; charset=utf-8');" + body + "}", 1},
		{"parenthesised concatenation", "function f($t){" + set + "header((('X-Tag: ') . $t));" + body + "}", 1},
		{"concatenated location", "function f($u){" + set + "header('Location: ' . $u);" + body + "}", 0},
		{"concatenated field name", "function f($n){" + set + "header('X-' . $n);" + body + "}", 0},
		{"dynamic leading operand", "function f($n){" + set + "header($n . ': 1');" + body + "}", 0},

		{"explicit other status", set + "header('X-Reason: fallback', true, 500);" + body, 0},
		{"named explicit status", set + "header('X-Reason: fallback', response_code: 500);" + body, 0},
		{"explicit status repeating 204", set + "header('X-Reason: same', true, 204);" + body, 1},
		{"explicit zero status", set + "header('X-Reason: same', true, 0);" + body, 1},
		{"explicit zero with location", set + "header('Location: /next', true, 0);" + body, 0},
		{"dynamic explicit status", "function f($c){" + set + "header('X-Reason: same', true, $c);" + body + "}", 0},

		{"header before status", "header('HTTP/1.1 200 OK');" + set + body, 1},
		{"header after body", set + body + "header('HTTP/1.1 200 OK');", 1},
		{"conditional header", "function f($c){" + set + "if($c){header('HTTP/1.1 200 OK');}" + body + "}", 0},
		{"header in closure", set + "$g = function(){header('HTTP/1.1 200 OK');};" + body, 1},
		{"other builtin call", set + "$n = strlen('abc');" + body, 1},
		{"user function shadowing header", "namespace App; function header($h){} \\http_response_code(204); header('HTTP/1.1 200 OK'); echo 'body';", 0},

		{"while repeats body", "function f($c){" + set + "while($c){" + body + "header('HTTP/1.1 200 OK');}}", 0},
		{"do while repeats body", "function f($c){" + set + "do{" + body + "header('HTTP/1.1 200 OK');}while($c);}", 0},
		{"for repeats body", set + "for($i=0;$i<2;$i++){" + body + "header('HTTP/1.1 200 OK');}", 0},
		{"foreach repeats body", set + "foreach([1,2] as $_){if($_){" + body + "}header('HTTP/1.1 200 OK');}", 0},
		{"loop with harmless header", set + "foreach([1,2] as $_){" + body + "header('X-Seen: 1');}", 1},
		{"header after loop", set + "foreach([1,2] as $_){" + body + "}header('HTTP/1.1 200 OK');", 1},
		{"loop outside function", "foreach([1] as $_){$g=function(){" + set + body + "};header('HTTP/1.1 200 OK');}", 1},

		{"open buffer", set + "ob_start();" + body, 0},
		{"buffer opened before status", "ob_start();" + set + body, 0},
		{"discarded buffer", set + "ob_start();" + body + "ob_end_clean();", 0},
		{"captured buffer", set + "ob_start();" + body + "$c = ob_get_clean();", 0},
		{"body after discarded buffer", set + "ob_start(); echo 'debug'; ob_end_clean();" + body, 1},
		{"body after flushed buffer", set + "ob_start(); ob_end_flush();" + body, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NoContentResponseWithBody"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
