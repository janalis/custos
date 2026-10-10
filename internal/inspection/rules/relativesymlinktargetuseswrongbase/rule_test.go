package relativesymlinktargetuseswrongbase

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
		{"<?php\nchdir('/tmp'); symlink('../data.txt', '/srv/links/current'); file_get_contents(readlink('/srv/links/current'));\n", 1},
		{"<?php\nchdir('/srv/links'); symlink('../data.txt', '/srv/links/current'); file_get_contents(readlink('/srv/links/current'));\n", 0},
		{"<?php other();", 0},
		{"<?php file_get_contents(readlink($path));", 0},
		{"<?php chdir('/tmp');symlink('/data','/srv/link');file_get_contents(readlink('/srv/link'));", 0},
		{"<?php symlink('../data','/srv/link');file_get_contents(readlink('/srv/link'));", 0},
		{"<?php chdir('/tmp');symlink('../data','/srv/link');unlink('/srv/link');file_get_contents(readlink('/srv/link'));", 0},
		{"<?php chdir('/tmp');symlink('../data','/srv/link');unknown();file_get_contents(readlink('/srv/link'));", 0},
		{"<?php $x=1;chdir('/tmp');symlink('../data','/srv/link');file_get_contents(readlink('/srv/link'));", 1},
		{"<?php if($ok){chdir('/tmp');symlink('../data','/srv/link');}file_get_contents(readlink('/srv/link'));", 0},
		{"<?php function unreachable(){return;\nchdir('/tmp'); symlink('../data.txt', '/srv/links/current'); file_get_contents(readlink('/srv/links/current'));\n}", 0},
		{"<?php \nchdir('/tmp'); symlink('../data.txt', '/srv/links/current'); file_get_contents(readlink('/srv/links/current'));\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RelativeSymlinkTargetUsesWrongBase"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/RelativeSymlinkTargetUsesWrongBase/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RelativeSymlinkTargetUsesWrongBase"}})
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
