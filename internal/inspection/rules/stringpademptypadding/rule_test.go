package stringpademptypadding

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
		{"<?php\nstr_pad('q', 8, '');\n", 1},
		{"<?php\nstr_pad('q', 1, '');\n", 0},
		{"<?php other();", 0},
		{"<?php str_pad(\"a\",$n,\"\");", 0},
		{"<?php str_pad($s,4,\"\");", 0},
		{"<?php str_pad(\"a\",4,$pad);", 0},
		{"<?php str_pad(\"a\",4,\"x\");", 0},
		{"<?php function unreachable(){return;\nstr_pad('q', 8, '');\n}", 0},
		{"<?php \nstr_pad('q', 8, '');\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"StringPadEmptyPadding"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
