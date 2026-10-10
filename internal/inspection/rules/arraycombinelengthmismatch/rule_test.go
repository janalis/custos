package arraycombinelengthmismatch

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
		{"array_combine([\"a\",\"b\"],[1]);", 1},
		{"array_combine([\"a\"],[1]);", 0},
		{"array_combine([\"a\"=>1,\"a\"=>2],[3]);", 0},
		{"array_combine($x,[1]);", 0},
		{"array_combine([$unknown=>1],[1,2]);", 0},
		{"array_combine(keys:[\"a\"],values:[1,2]);", 1},
		{"array_combine([PHP_INT_MAX=>1,2],[3]);", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayCombineLengthMismatch"}})
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
