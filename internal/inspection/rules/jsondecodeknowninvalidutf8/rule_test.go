package jsondecodeknowninvalidutf8

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
		{"<?php\njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n", 1},
		{"<?php\njson_decode('{}', true, 512, JSON_THROW_ON_ERROR);\n", 0},
		{"<?php other();", 0},
		{"<?php json_decode(\"\\xFF\",true,512,$flags);", 0},
		{"<?php json_decode(\"\\xFF\",true,512,JSON_INVALID_UTF8_SUBSTITUTE);", 0},
		{"<?php json_decode(\"\\xFF\");", 1},
		{"<?php function unreachable(){return;\njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n}", 0},
		{"<?php \njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonDecodeKnownInvalidUtf8"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/JsonDecodeKnownInvalidUtf8/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonDecodeKnownInvalidUtf8"}})
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
