package logarithminvalidrealdomain

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
		{"log10(0);", 1},
		{"log1p(-1);", 1},
		{"log1p(-0.5);", 0},
		{"log(2,1);", 1},
		{"log(2,0);", 1},
		{"log(2,2);", 0},
		{"log(+0.5);", 0},
		{"log(-0.5);", 1},
		{"log($x);", 0},
		{"log(!$x);", 0},
		{"log('2');", 0},
		{"sin(2);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"LogarithmInvalidRealDomain"}})
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
