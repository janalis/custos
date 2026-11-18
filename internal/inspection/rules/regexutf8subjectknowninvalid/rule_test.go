package regexutf8subjectknowninvalid

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
		{"<?php\npreg_match('~.+~u', \"\\xFF\");\n", 1},
		{"<?php\npreg_match('~.+~', \"\\xFF\");\n", 0},
		{"<?php other();", 0},
		{"<?php preg_replace('~.~u','x',\"\\xFF\");", 1},
		{"<?php preg_match($p,$s);", 0},
		{"<?php preg_match('x',$s);", 0},
		{"<?php preg_match('~abc',\"\\xFF\");", 0},
		{"<?php preg_split('~.~u',\"\\xFF\");", 1},
		{"<?php function unreachable(){return;\npreg_match('~.+~u', \"\\xFF\");\n}", 0},
		{"<?php \npreg_match('~.+~u', \"\\xFF\");\n function broken(", 0},
		// Bracket-style delimiters close on their counterpart, not on the opener.
		{"<?php preg_match('[a[u]]', \"u\\xFF\");", 0},
		{"<?php preg_match('{a{1}u}', \"au\\xFF\");", 0},
		{"<?php preg_match('(a(u))', \"au\\xFF\");", 0},
		{"<?php preg_match('<a<u>>', \"au\\xFF\");", 0},
		{"<?php preg_match('(a(u))u', \"au\\xFF\");", 1},
		{"<?php preg_match('{a{1}}iu', \"a\\xFF\");", 1},
		{"<?php preg_match('[a]u', \"a\\xFF\");", 1},
		{"<?php preg_match('(a(u)', \"au\\xFF\");", 0},
		// The body ends at the first unescaped delimiter.
		{"<?php preg_match('~a~x~u', \"a\\xFF\");", 0},
		{"<?php preg_match('~a\\\\~b~u', \"a\\xFF\");", 1},
		{"<?php preg_match('~a\\\\', \"a\\xFF\");", 0},
		// Only modifier letters may follow the body.
		{"<?php preg_match('~a~ui', \"a\\xFF\");", 1},
		{"<?php preg_match('~a~u ', \"a\\xFF\");", 1},
		{"<?php preg_match('~a~uq', \"a\\xFF\");", 0},
		{"<?php preg_match('~a~i', \"a\\xFF\");", 0},
		// Leading whitespace is skipped; invalid delimiters never compile.
		{"<?php preg_match(' ~a~u', \"a\\xFF\");", 1},
		{"<?php preg_match('', \"a\\xFF\");", 0},
		{"<?php preg_match('   ', \"a\\xFF\");", 0},
		{"<?php preg_match('uau', \"a\\xFF\");", 0},
		{"<?php preg_match('1a1u', \"a\\xFF\");", 0},
		{"<?php preg_match('AaAu', \"a\\xFF\");", 0},
		{"<?php preg_match('\\\\a\\\\u', \"a\\xFF\");", 0},
		{"<?php preg_match(\"\\0a\\0u\", \"a\\xFF\");", 0},
		{"<?php preg_match(\"\\xE9a\\xE9u\", \"a\\xFF\");", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexUtf8SubjectKnownInvalid"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
