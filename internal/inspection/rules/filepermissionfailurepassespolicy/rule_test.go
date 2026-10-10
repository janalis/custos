package filepermissionfailurepassespolicy

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
		{"<?php\nif ((fileperms($path) & 0002) === 0) { return true; }\n", 1},
		{"<?php\n$m = fileperms($path); if ($m !== false && ($m & 0002) === 0) { return true; }\n", 0},
		{"<?php other();", 0},
		{"<?php if((fileperms($p)&$mask)===0){return true;}", 0},
		{"<?php if((fileperms($p)&0)===0){return true;}", 0},
		{"<?php if(($mode&2)===0){return true;}", 0},
		{"<?php if((fileperms($p)&2)===0){throw new Exception;}", 0},
		{"<?php if((fileperms($p)&2)===0){echo 'ok';}", 0},
		{"<?php $ok=(fileperms($p)&2)===0;", 0},
		{"<?php if((fileperms($p)&2)==0){return true;}", 1},
		{"<?php if((fileperms($p)&2)===1){return true;}", 0},
		{"<?php function unreachable(){return;\nif ((fileperms($path) & 0002) === 0) { return true; }\n}", 0},
		{"<?php \nif ((fileperms($path) & 0002) === 0) { return true; }\n function broken(", 0},
		{"<?php $r=fileperms($p);$alias=&$r;mutate($alias);if(($r&2)===0){return true;}", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/FilePermissionFailurePassesPolicy/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
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
