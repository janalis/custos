package arrayfillnegativecount

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
		{"<?php\narray_fill(0, -4, 'q');\n", 1},
		{"<?php\narray_fill(0, 4, 'q');\n", 0},
		{"<?php other();", 0},
		{"<?php array_fill(0,$n,0);", 0},
		{"<?php array_fill(count:-1,start_index:0,value:0);", 1},
		{"<?php function unreachable(){return;\narray_fill(0, -4, 'q');\n}", 0},
		{"<?php \narray_fill(0, -4, 'q');\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayFillNegativeCount"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
