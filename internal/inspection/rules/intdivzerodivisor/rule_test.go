package intdivzerodivisor

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nintdiv(14, 0);\n", 1},
		{"<?php\nintdiv(14, 2);\n", 0},
		{"<?php other();", 0},
		{"<?php intdiv(12,$n);", 0},
		{"<?php intdiv(num2:0,num1:12);", 1},
		{"<?php function unreachable(){return;\nintdiv(14, 0);\n}", 0},
		{"<?php \nintdiv(14, 0);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IntdivZeroDivisor"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
