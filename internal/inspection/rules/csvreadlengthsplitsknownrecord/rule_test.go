package csvreadlengthsplitsknownrecord

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
		{"<?php\n$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 4);\n", 1},
		{"<?php\n$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 100);\n", 0},
		{"<?php other();", 0},
		{"<?php fgetcsv($h,4);", 0},
		{"<?php $h=fopen('local.csv','r');fgetcsv($h,4);", 0},
		{"<?php $h=fopen('data://text/plain,abc%ZZ','r');fgetcsv($h,1);", 0},
		{"<?php $h=fopen('data://text/plain,abc','w');fgetcsv($h,1);", 0},
		{"<?php $h=fopen('data://text/plain,%22abcdefgh%22,ij%0A','r');fgetcsv($h,4);", 0},
		{"<?php $h=fopen('data://text/plain,abcdefgh,ij%0A','r');fgetcsv($h,4,';');", 0},
		{"<?php $h=fopen('data://text/plain,abcdefgh,ij%0A','r');fgets($h);fgetcsv($h,4);", 0},
		{"<?php $h=fopen('data://text/plain,abcdefgh,ij','rb');fgetcsv($h,4,',','\"','\\\\');", 1},
		{"<?php $h=fopen('data://text/plain,abc','r');fgetcsv($h,0);", 0},
		{"<?php function unreachable(){return;\n$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 4);\n}", 0},
		{"<?php \n$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 4);\n function broken(", 0},
		{"<?php $h=fopen('data://text/plain,abcdef,ij','r');$alias=&$h;mutate($alias);fgetcsv($h,4);", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CsvReadLengthSplitsKnownRecord"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/CsvReadLengthSplitsKnownRecord/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CsvReadLengthSplitsKnownRecord"}})
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
