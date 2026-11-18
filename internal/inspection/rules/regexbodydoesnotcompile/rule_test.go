package regexbodydoesnotcompile

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
		{"<?php\npreg_match('~[~', $text);\n", 1},
		{"<?php\npreg_match('~[a-z]~', $text);\n", 0},
		{"<?php other();", 0},
		{"<?php preg_match('x',$s);", 0},
		{"<?php preg_match('/abc',$s);", 0},
		{"<?php preg_match('~(?=a)(~',$s);", 0},
		{"<?php preg_match('~(*SKIP)(~',$s);", 0},
		{"<?php preg_match('~(~x',$s);", 0},
		{"<?php preg_match('~\\\\Q(~',$s);", 0},
		{"<?php preg_match('~a~b~',$s);", 0},
		{"<?php preg_match('~[[:alpha:]]~',$s);", 0},
		{"<?php preg_match('~[]a]~',$s);", 0},
		{"<?php preg_match('~[^]a]~',$s);", 0},
		{"<?php preg_match('~a)~',$s);", 1},
		{"<?php preg_match('~(a~',$s);", 1},
		{"<?php preg_match('~(a)~',$s);", 0},
		{"<?php preg_match('~a\\\\~',$s);", 0},
		{"<?php preg_match('~a\\\\]~',$s);", 0},
		{"<?php function unreachable(){return;\npreg_match('~[~', $text);\n}", 0},
		{"<?php \npreg_match('~[~', $text);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexBodyDoesNotCompile"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}

func BenchmarkContracts(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexBodyDoesNotCompile"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php preg_match('~(abc~',$text);`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}
