package stringoffsetwritetruncatesvalue

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
		{"<?php\n$s = 'dog'; $s[0] = 'AB';\n", 1},
		{"<?php\n$s = 'dog'; $s[0] = 'A';\n", 0},
		{"<?php other();", 0},
		{"<?php $s=\"abc\";$s[0].=\"AB\";", 0},
		{"<?php $s=[\"a\"];$s[0]=\"AB\";", 0},
		{"<?php $s=\"abc\";$s[$i]=\"AB\";", 0},
		{"<?php $s=\"abc\";$s[0]=$v;", 0},
		{"<?php $s=\"abc\";$t=\"AB\";$s[0]=&$t;", 0},
		{"<?php $s=\"AB\";", 0},
		{"<?php function unreachable(){return;\n$s = 'dog'; $s[0] = 'AB';\n}", 0},
		{"<?php \n$s = 'dog'; $s[0] = 'AB';\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"StringOffsetWriteTruncatesValue"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
