package gmpdivisionpairusedasnumber

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
		{"gmp_add('2',gmp_div_qr('3','2'));", 1},
		{"$p=gmp_div_qr('3','2');gmp_strval($p);", 1},
		{"gmp_strval(gmp_div_qr('3','2')[0]);", 0},
		{"gmp_abs(gmp_div_qr('3','2'));", 1},
		{"strlen('2');", 0},
		{"gmp_strval($x);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GmpDivisionPairUsedAsNumber"}})
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
