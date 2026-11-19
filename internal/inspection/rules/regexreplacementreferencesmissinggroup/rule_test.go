package regexreplacementreferencesmissinggroup

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
		{"<?php\npreg_replace('~([a-z]+)~', '$2', $text);\n", 1},
		{"<?php\npreg_replace('~([a-z]+)~', '$1', $text);\n", 0},
		{"<?php other();", 0},
		{"<?php preg_replace($p,'$2',$s);", 0},
		{"<?php preg_replace('~(?|a(b)|c(d))~','$2',$s);", 0},
		{"<?php preg_replace('~(a)~','abc',$s);", 0},
		{"<?php preg_replace('~(a)~','${2}',$s);", 1},
		{"<?php preg_replace('~(a)~','${2',$s);", 0},
		{"<?php preg_replace('~(a)~','\\\\$2',$s);", 0},
		{"<?php preg_replace('~(a)~','\\2',$s);", 1},
		{"<?php preg_replace('~(a)~','x\\\\',$s);", 0},
		{"<?php preg_replace('~(a)~','$1-$0',$s);", 0},
		{"<?php preg_replace('~(a)~','\\\\\\\\\\\\\\\\2',$s);", 0},
		{"<?php function unreachable(){return;\npreg_replace('~([a-z]+)~', '$2', $text);\n}", 0},
		{"<?php \npreg_replace('~([a-z]+)~', '$2', $text);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexReplacementReferencesMissingGroup"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
