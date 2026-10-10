package regexutf8subjectknowninvalid

import (
	"os"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
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

func TestFixtureRanges(t *testing.T) {
	for _, name := range []string{"basic", "negative"} {
		marked, err := os.ReadFile("../../../../testdata/rules/RegexUtf8SubjectKnownInvalid/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexUtf8SubjectKnownInvalid"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", src, syntax.Options{}))
		if len(got) != len(want) {
			t.Fatalf("%s: got %+v, want %+v", name, got, want)
		}
		for i, g := range got {
			w := want[i]
			if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity || len(g.Fixes) != 0 {
				t.Fatalf("got %+v, want %+v", g, w)
			}
		}
	}
}
