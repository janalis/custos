package arraychunkinvalidsize

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
		{"array_chunk($items,0);", 1},
		{"array_chunk($items,-1);", 1},
		{"array_chunk($items,+1);", 0},
		{"array_chunk($items,1);", 0},
		{"array_chunk($items,$n);", 0},
		{"array_chunk(length:0,array:$items);", 1},
		{"array_chunk($items,0.5);", 0},
		{"array_chunk($items,!1);", 0},
		{"array_chunk($items,-$n);", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayChunkInvalidSize"}})
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
