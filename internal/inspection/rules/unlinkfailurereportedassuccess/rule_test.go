package unlinkfailurereportedassuccess

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
		{"<?php\nfunction removeFile($p) { unlink($p); return true; }\n", 1},
		{"<?php\nfunction removeFile($p) { return unlink($p); }\n", 0},
		{"<?php other();", 0},
		{"<?php unlink($p);return true;", 0},
		{"<?php function f($p){unlink($p);return false;}", 0},
		{"<?php function unreachable(){return;unlink($p);return true;}", 0},
		{"<?php \nfunction removeFile($p) { unlink($p); return true; }\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnlinkFailureReportedAsSuccess"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
