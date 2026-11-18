package fractionalmodulooperand

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
		{"<?php\n$r = 9 % 0.25;\n", 1},
		{"<?php\n$r = fmod(9, 0.25);\n", 0},
		{"<?php other();", 0},
		{"<?php $x=9%2.0;", 0},
		{"<?php $x=9%-2.5;", 1},
		{"<?php $x=9%$v;", 0},
		{"<?php $x=9+0.25;", 0},
		{"<?php $x=9%1e999;", 0},
		{"<?php function unreachable(){return;\n$r = 9 % 0.25;\n}", 0},
		{"<?php \n$r = 9 % 0.25;\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FractionalModuloOperand"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
