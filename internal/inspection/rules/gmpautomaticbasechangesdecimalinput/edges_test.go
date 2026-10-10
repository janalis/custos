package gmpautomaticbasechangesdecimalinput

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
		{"gmp_init('024',0);", 1},
		{"gmp_init(num:'024');", 1},
		{"gmp_init('024',base:0);", 1},
		{"gmp_init('024',10);", 0},
		{"gmp_init('0007');", 0},
		{"gmp_init('0000');", 0},
		{"gmp_init('09');", 0},
		{"gmp_init('0x20');", 0},
		{"gmp_init('2');", 0},
		{"gmp_init($s);", 0},
		{"gmp_init('024',$base);", 0},
		{"gmp_init('024',1-1);", 0},
		{"gmp_init('024' /* retain */);", 1},
		{"strlen('024');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GmpAutomaticBaseChangesDecimalInput"}})
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
