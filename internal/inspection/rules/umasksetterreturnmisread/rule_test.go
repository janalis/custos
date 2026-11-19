package umasksetterreturnmisread

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
		{"<?php\nif (umask(0077) !== 0077) { throw new RuntimeException(); }\n", 1},
		{"<?php\numask(0077); if (umask() !== 0077) { throw new RuntimeException(); }\n", 0},
		{"<?php other();", 0},
		{"<?php if(umask($mask)!==7){return false;}", 0},
		{"<?php if(umask(7)!==0){return false;}", 0},
		{"<?php if(umask(0)!==0){return false;}", 0},
		{"<?php if(umask(7)!==7){return false;}", 1},
		{"<?php $v=umask(7)!==7;", 0},
		{"<?php if(umask(7)!==7){echo 1;}", 0},
		{"<?php if(umask(7)==7){return false;}", 0},
		{"<?php if($v!==7){return false;}", 0},
		{"<?php function unreachable(){return;\nif (umask(0077) !== 0077) { throw new RuntimeException(); }\n}", 0},
		{"<?php \nif (umask(0077) !== 0077) { throw new RuntimeException(); }\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UmaskSetterReturnMisread"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
