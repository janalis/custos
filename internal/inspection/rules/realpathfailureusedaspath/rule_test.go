package realpathfailureusedaspath

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
		{"<?php\n$root = realpath($path); file_get_contents($root . '/config.json');\n", 1},
		{"<?php\n$root = realpath($path); if ($root === false) { throw new RuntimeException(); } file_get_contents($root . '/config.json');\n", 0},
		{"<?php other();", 0},
		{"<?php file_get_contents($path);", 0},
		{"<?php file_get_contents('x'.realpath($path));", 1},
		{"<?php $r=realpath($p);$r='/safe';file_get_contents($r.'/x');", 0},
		{"<?php function unreachable(){return;\n$root = realpath($path); file_get_contents($root . '/config.json');\n}", 0},
		{"<?php \n$root = realpath($path); file_get_contents($root . '/config.json');\n function broken(", 0},
		{"<?php $r=realpath($p);$alias=&$r;mutate($alias);file_get_contents($r.'/x');", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RealpathFailureUsedAsPath"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/RealpathFailureUsedAsPath/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RealpathFailureUsedAsPath"}})
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
