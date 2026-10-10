package arrayflipduplicatevalueloss

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRule(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"array_flip([\"a\"=>\"red\",\"b\"=>\"red\"]);", 1},
		{"array_flip([1,\"1\"]);", 1},
		{"array_flip([\"red\",\"blue\"]);", 0},
		{"array_flip([true,false]);", 0},
		{"array_flip([\"a\"=>\"red\",\"a\"=>\"red\"]);", 0},
		{"array_flip([$x,$y]);", 0},
		{"array_flip($x);", 0},
		{"array_flip([...$x]);", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayFlipDuplicateValueLoss"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
			for _, g := range got {
				if g.Message != message || g.Span.End <= g.Span.Start {
					t.Fatalf("bad finding: %+v", g)
				}
			}
		})
	}
}
