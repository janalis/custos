package contentlengthknownbodymismatch

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestEmptyExitArgumentList covers paths closed by exit() or die() with an
// empty argument list: they emit nothing and complete the response.
func TestEmptyExitArgumentList(t *testing.T) {
	deep := "str_repeat('a',1)" + strings.Repeat(".'a'", 10)
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"header('Content-Length: 9'); echo 'a'; exit();", 1},
		{"header('Content-Length: 9'); echo 'a'; die();", 1},
		{"(header('Content-Length: 9')); echo 'a'; (exit());", 1},
		{"namespace A; header('Content-Length: 9'); echo 'a'; exit();", 1},
		{"namespace A { header('Content-Length: 9'); echo 'a'; exit(); }", 1},
		{"function x(){header('Content-Length: 9');echo 'a';exit();}", 1},
		{"function x(){header('Content-Length: 9');echo 'a';die();}", 1},
		{"header('Content-Length: 9'); print 'ab'; exit();", 1},
		{"header('Content-Length: 9'); echo 'a' . str_repeat('x', 2), 'b'; exit();", 1},
		{"header('X-A: b', false); header('Content-Length: 9'); echo 'a'; exit();", 1},
		{"header('Content-Length: 0'); exit();", 0},
		{"header('Content-Length: 1'); echo 'a'; exit();", 0},
		{"header('Content-Length: 4'); echo 'a' . str_repeat('x', 2), 'b'; exit();", 0},
		{"header('Content-Length: 9'); echo $unknown; exit();", 0},
		{"header('Content-Length: 9'); print $unknown; exit();", 0},
		{"header('Content-Length: 9'); echo 'a'; exit($unknown);", 0},
		{"header('Content-Length: 9'); echo 'a' - 'b'; exit();", 0},
		{"header('Content-Length: 9'); echo strtolower($a); exit();", 0},
		{"header('Content-Length: 9'); echo str_repeat('x', $n); exit();", 0},
		{"header('Content-Length: 9'); echo str_repeat('x', -1); exit();", 0},
		{"header('Content-Length: 9'); echo str_repeat($s, 2); exit();", 0},
		{"header('Content-Length: 9'); echo str_repeat('xx', 1073741824); exit();", 0},
		{"header('Content-Length: 9'); echo str_repeat('', 3) . str_repeat('x', 0); exit();", 1},
		{"header('Content-Length: 9'); echo " + deep + "; exit();", 0},
		{"header('Content-Length: 9'); unknown(); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); $x = 1; echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); if ($x) {} echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header($dynamic); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header('Location: /x'); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header('Status: 204'); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header('Transfer-Encoding: chunked'); echo 'a'; exit();", 0},
		{"header('Content-Length: 9', true, 204); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header('X-A: b', true, 304); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); http_response_code(204); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header_remove('Content-Length'); echo 'a'; exit();", 0},
		{"header('Content-Length: 9'); header('Content-Length: 1'); echo 'a'; exit();", 0},
		{"$x = 1; header('Content-Length: 9'); echo 'a'; exit();", 0},
		{"if ($x) {} header('Content-Length: 9'); echo 'a'; exit();", 0},
		{"unknown(); header('Content-Length: 9'); echo 'a'; exit();", 0},
		{"function f() {} header('Content-Length: 9'); echo 'a'; exit();", 0},
		{"if ($x) header('Content-Length: 9'); echo 'a'; exit();", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthKnownBodyMismatch"}})
			if err != nil {
				t.Fatal(err)
			}
			file := syntax.Parse("review.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(file.Errors) != 0 {
				t.Fatal(file.Errors)
			}
			got := e.Analyze(file)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
