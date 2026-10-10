package regexreplacementcallbackfallsthrough

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
		{"<?php\npreg_replace_callback('~[a-z]+~', function($m) { strtoupper($m[0]); }, $text);\n", 1},
		{"<?php\npreg_replace_callback('~[a-z]+~', function($m) { return strtoupper($m[0]); }, $text);\n", 0},
		{"<?php other();", 0},
		{"<?php preg_replace_callback('~.~',$cb,$s);", 0},
		{"<?php preg_replace_callback('~.~',fn($m)=>$m[0],$s);", 0},
		{"<?php preg_replace_callback('~.~',function($m){yield $m[0];},$s);", 0},
		{"<?php preg_replace_callback('~.~',function($m){while($ok){}},$s);", 0},
		{"<?php preg_replace_callback('~.~',function($m){$cb=function(){return 1;};},$s);", 1},
		{"<?php preg_replace_callback('~.~',function($m){if($ok){return 'a';}},$s);", 1},
		{"<?php function unreachable(){return;\npreg_replace_callback('~[a-z]+~', function($m) { strtoupper($m[0]); }, $text);\n}", 0},
		{"<?php \npreg_replace_callback('~[a-z]+~', function($m) { strtoupper($m[0]); }, $text);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexReplacementCallbackFallsThrough"}})
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
	for _, name := range []string{"basic", "negative", "audit-negative-2"} {
		marked, err := os.ReadFile("../../../../testdata/rules/RegexReplacementCallbackFallsThrough/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RegexReplacementCallbackFallsThrough"}})
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
