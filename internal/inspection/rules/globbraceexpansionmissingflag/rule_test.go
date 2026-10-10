package globbraceexpansionmissingflag

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
		{"<?php\nglob('src/*.{php,phtml}');\n", 1},
		{"<?php\nglob('src/*.{php,phtml}', GLOB_BRACE);\n", 0},
		{"<?php other();", 0},
		{"<?php glob($p);", 0},
		{"<?php glob('a.{b,c}',$flags);", 0},
		{"<?php glob('a.{b}');", 0},
		{"<?php glob('a.{b,c}',0);", 1},
		{"<?php glob('a.{b,c}',GLOB_NOSORT);", 1},
		{"<?php glob('a.\\\\{b,c}');", 0},
		{"<?php glob('a,b}');", 0},
		{"<?php function unreachable(){return;\nglob('src/*.{php,phtml}');\n}", 0},
		{"<?php \nglob('src/*.{php,phtml}');\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GlobBraceExpansionMissingFlag"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/GlobBraceExpansionMissingFlag/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GlobBraceExpansionMissingFlag"}})
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
