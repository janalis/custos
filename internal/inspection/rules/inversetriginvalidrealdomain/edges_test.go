package inversetriginvalidrealdomain

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
		{"acos(-1.5);", 1},
		{"acos(1);", 0},
		{"asin(-1);", 0},
		{"asin(+1.5);", 1},
		{"asin(0.0);", 0},
		{"asin(!$x);", 0},
		{"asin('2');", 0},
		{"asin($x);", 0},
		{"cos(2);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"InverseTrigInvalidRealDomain"}})
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
