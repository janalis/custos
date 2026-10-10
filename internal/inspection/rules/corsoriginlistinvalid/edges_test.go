package corsoriginlistinvalid

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
		{"header('Access-Control-Allow-Origin: https://a.test');", 0},
		{"header('Access-Control-Allow-Origin: https://a.test https://b.test');", 1},
		{"header('Access-Control-Allow-Origin: null https://a.test');", 0},
		{"header('Access-Control-Allow-Origin: https://a.test/ https://b.test');", 0},
		{"header('Access-Control-Allow-Origin: http://%xx http://a.test');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CorsOriginListInvalid"}})
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
