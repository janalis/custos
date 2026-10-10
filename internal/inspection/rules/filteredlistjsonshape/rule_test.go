package filteredlistjsonshape

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
		{"json_encode(array_filter([1, 0, 2]));", 1},
		{"json_encode(array_filter([1, 0, 2], null));", 1},
		{"json_encode(array_filter([1, 2, 0]));", 0},
		{"json_encode(array_filter([0, 1]));", 1},
		{"json_encode(array_filter([1, $x]));", 0},
		{"json_encode(array_filter([2 => 1, 3 => 0]));", 0},
		{"json_encode(array_filter([0 => 1, 1 => 0, 2 => 2]));", 1},
		{"json_encode(array_values(array_filter([0,1])));", 0},
		{"json_encode(array_filter([0,1], fn($x)=>true));", 0},
		{"json_encode($x);", 0},
		{"json_encode(array_filter($x));", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilteredListJsonShape"}})
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
