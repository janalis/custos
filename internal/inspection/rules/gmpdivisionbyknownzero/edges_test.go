package gmpdivisionbyknownzero

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
		{"gmp_mod(2,'+-0');", 0},
		{"gmp_mod('2',0);", 1},
		{"gmp_div_r('2','+000');", 1},
		{"gmp_div_qr('2','-0');", 1},
		{"gmp_div_q('2',gmp_init('0'));", 1},
		{"gmp_div_q('2',gmp_init(0,10));", 1},
		{"gmp_div_q('2',gmp_init('0',63));", 0},
		{"gmp_div_q('2',gmp_init('0',$base));", 0},
		{"gmp_div_q('2',gmp_init(gmp_init('0')));", 0},
		{"gmp_div_q('2','');", 0},
		{"gmp_div_q('2','0x0');", 0},
		{"gmp_div_q('2','2');", 0},
		{"gmp_div_q('2',$x);", 0},
		{"strlen('0');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GmpDivisionByKnownZero"}})
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
