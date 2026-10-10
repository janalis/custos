package bcmathexponentnotation

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
		{"bcsub('1','-2.5E-4');", 1},
		{"bcadd('e3','1');", 0},
		{"bcadd('3E+2','1');", 1},
		{"bcadd($unknown,'1');", 0},
		{"bcsqrt('4e2');", 1},
		{"bcpow('4e2',2);", 1},
		{"bcpowmod('4e2','2','3');", 1},
		{"bcadd(...$args);", 0},
		{"strlen('1e3');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BcMathExponentNotation"}})
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
