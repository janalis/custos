package arraymergenumerickeyloss

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
		{"array_merge([42=>\"a\"],[87=>\"b\"]);", 1},
		{"array_merge([\"a\"=>1],[\"b\"=>2]);", 0},
		{"array_merge([0=>\"a\",1=>\"b\"]);", 0},
		{"array_merge($unknown);", 0},
		{"array_merge([...$unknown]);", 0},
		{"array_merge(...$unknown);", 0},
		{"array_merge([\"42\"=>\"a\"]);", 1},
		{"array_merge([\"042\"=>\"a\"]);", 0},
		{"array_merge([$key=>\"a\"]);", 0},
		{"array_merge([\"a\"]);", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMergeNumericKeyLoss"}})
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
