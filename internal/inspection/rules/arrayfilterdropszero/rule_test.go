package arrayfilterdropszero

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
		{"array_filter([0, 2]);", 1},
		{"array_filter([0]);", 0},
		{"array_filter([false, 2]);", 0},
		{"array_filter([0, 2], fn($v) => true);", 0},
		{"array_filter($unknown);", 0},
		{"array_filter([\"0\", \"a\"]);", 1},
		{"array_filter([0, false]);", 0},
		{"array_filter([0, \"a\"], callback: null);", 1},
		{"array_filter([0=>0,0=>2,1=>3]);", 0},
		{"array_filter(array: [0, \"a\"]);", 1},
		{"array_filter(...$args);", 0},
		{"other([0,1]);", 0},
		{"namespace Custom; function array_filter($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayFilterDropsZero"}})
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
