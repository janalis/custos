package transferencodingwithcontentlength

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
		{"header('Content-Length: 1');", 0},
		{"header('Transfer-Encoding: chunked');header('Content-Length: 1');", 1},
		{"header('Content-Length: 1');header_remove('Content-Length');header('Transfer-Encoding: chunked');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"TransferEncodingWithContentLength"}})
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
