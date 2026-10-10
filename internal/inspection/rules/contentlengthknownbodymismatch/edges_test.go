package contentlengthknownbodymismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"strlen('x');", 0},
		{"header('X: y');", 0},
		{"header('Content-Length: nope');", 0},
		{"header('Content-Length: 3'); echo 'abc';", 0},
		{"header('Content-Length: 9'); echo $unknown;", 0},
		{"header('Content-Length: 9'); header_remove(); echo 'a';", 0},
		{"header('Content-Length: 9'); header('Content-Length: 1'); echo 'a';", 0},
		{"header('Content-Length: 9'); unknown(); echo 'a';", 0},
		{"header('Content-Length: 9'); http_response_code(204); echo 'a';", 0},
		{"header('Content-Length: 9'); header('Transfer-Encoding: chunked'); echo 'a';", 0},
		{"function x(){header('Content-Length: 9');echo 'a';}", 0},
		{"function x(){header('Content-Length: 9');echo 'a';exit;}", 1},
		{"header('Content-Length: 9'); print 'ab'; exit('c');", 1},
		{"header('Content-Length: 9'); echo str_repeat('x',2);", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthKnownBodyMismatch"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
