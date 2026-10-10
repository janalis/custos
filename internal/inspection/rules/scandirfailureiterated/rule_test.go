package scandirfailureiterated

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
		{"<?php\nforeach (scandir($directory) as $entry) { echo $entry; }\n", 1},
		{"<?php\n$entries = scandir($directory); if ($entries !== false) { foreach ($entries as $entry) { echo $entry; } }\n", 0},
		{"<?php other();", 0},
		{"<?php foreach($unknown as $x){}", 0},
		{"<?php $a=scandir($p);mutate($a);foreach($a as $x){}", 0},
		{"<?php $a=scandir($p);if($a===false){return;}foreach($a as $x){}", 0},
		{"<?php function unreachable(){return;\nforeach (scandir($directory) as $entry) { echo $entry; }\n}", 0},
		{"<?php \nforeach (scandir($directory) as $entry) { echo $entry; }\n function broken(", 0},
		{"<?php $a=scandir($p);$alias=&$a;mutate($alias);foreach($a as $entry){}", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ScandirFailureIterated"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/ScandirFailureIterated/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ScandirFailureIterated"}})
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
