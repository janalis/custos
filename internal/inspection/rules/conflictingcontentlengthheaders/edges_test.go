package conflictingcontentlengthheaders

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
		{"header('Content-Length: bad',false);", 0},
		{"header('Content-Length: 1');header('Content-Length: 1',false);", 0},
		{"header('Content-Length: bad');header('Content-Length: 2',false);", 0},
		{"header('Content-Length: 1');header('Content-Length: 2');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ConflictingContentLengthHeaders"}})
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
