package bcmathfloatoperand

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
		{"bcadd(0.5,'0');", 1},
		{"bcsub('0',0.5);", 1},
		{"bcmul('1',2);", 0},
		{"bcsqrt(0.5);", 1},
		{"bcpow(0.5,2);", 1},
		{"bcadd(num2:0.5,num1:'1');", 1},
		{"bcadd(...$args);", 0},
		{"bcadd();", 0},
		{"strlen('x');", 0},
		{"$x=0.5;$x='1';bcadd($x,'0');", 0},
		{"namespace A;function bcadd($a,$b){}bcadd(0.5,'1');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BcMathFloatOperand"}})
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
