package intdivminimumintegeroverflow

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
		{"<?php\nintdiv(PHP_INT_MIN, -1);\n", 1},
		{"<?php\nintdiv(PHP_INT_MIN, 1);\n", 0},
		{"<?php other();", 0},
		{"<?php intdiv(1,-1);", 0},
		{"<?php intdiv(PHP_INT_MIN,$n);", 0},
		{"<?php intdiv($n,-1);", 0},
		{"<?php function unreachable(){return;\nintdiv(PHP_INT_MIN, -1);\n}", 0},
		{"<?php \nintdiv(PHP_INT_MIN, -1);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IntdivMinimumIntegerOverflow"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
