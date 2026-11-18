package copyfailurereportedassuccess

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
		{"<?php\nfunction duplicate($a, $b) { copy($a, $b); return true; }\n", 1},
		{"<?php\nfunction duplicate($a, $b) { return copy($a, $b); }\n", 0},
		{"<?php other();", 0},
		{"<?php copy($a,$b);return true;", 0},
		{"<?php function f($a,$b){copy($a,$b);return false;}", 0},
		{"<?php function unreachable(){return;copy($a,$b);return true;}", 0},
		{"<?php \nfunction duplicate($a, $b) { copy($a, $b); return true; }\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CopyFailureReportedAsSuccess"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
