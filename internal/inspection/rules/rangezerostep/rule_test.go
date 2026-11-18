package rangezerostep

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nrange(3, 9, 0);\n", 1},
		{"<?php\nrange(3, 9, 2);\n", 0},
		{"<?php other();", 0},
		{"<?php range(3,9,0.0);", 1},
		{"<?php range(3,9,1.0);", 0},
		{"<?php range(3,9,$v);", 0},
		{"<?php range(3,9);", 0},
		{"<?php range(step:0,start:3,end:9);", 1},
		{"<?php function unreachable(){return;\nrange(3, 9, 0);\n}", 0},
		{"<?php \nrange(3, 9, 0);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RangeZeroStep"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
