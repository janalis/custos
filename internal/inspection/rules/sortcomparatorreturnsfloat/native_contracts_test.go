package sortcomparatorreturnsfloat

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"usort($xs, fn(float $a, float $b) => $a - $b);", 1},
		{"usort($xs, fn(int $a, int $b) => $a - $b);", 0},
		{"namespace Custom; function unrelated($x) {} unrelated(1);", 0},
		{"usort($xs, $unknown); array_map($unknown, $xs);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SortComparatorReturnsFloat"}})
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
					t.Fatalf("invalid finding: %+v", g)
				}
			}
		})
	}
}
