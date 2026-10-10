package bcmathscalediscardsrequiredfraction

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
		{"bcsub('1.2','1.1',0);", 1},
		{"bcmul('0.25','0.5',2);", 1},
		{"bcadd('1.5','0.5',0);", 0},
		{"bcadd('1.20','0',1);", 0},
		{"bcdiv('1.2','1',0);", 0},
		{"bcadd('1.2','1');", 0},
		{"bcadd('1.2','1',-1);", 0},
		{"bcadd('1.2','1',1025);", 0},
		{"bcadd($x,'1',0);", 0},
		{"bcadd('1.2',$x,0);", 0},
		{"bcadd('bad','1',0);", 0},
		{"bcadd('1','bad',0);", 0},
		{"bcadd(str_repeat('1',1025),'0',0);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BcMathScaleDiscardsRequiredFraction"}})
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
