package negativebitshiftcount

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
		{"<?php\n$mask = 2 << -3;\n", 1},
		{"<?php\n$mask = 2 << 3;\n", 0},
		{"<?php other();", 0},
		{"<?php $x=1>>-1;", 1},
		{"<?php $x=1<<$n;", 0},
		{"<?php $x=1+2;", 0},
		{"<?php function unreachable(){return;\n$mask = 2 << -3;\n}", 0},
		{"<?php \n$mask = 2 << -3;\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NegativeBitShiftCount"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
