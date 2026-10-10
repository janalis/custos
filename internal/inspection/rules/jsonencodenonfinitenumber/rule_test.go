package jsonencodenonfinitenumber

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
		{"<?php\njson_encode(['amount' => INF], JSON_THROW_ON_ERROR);\n", 1},
		{"<?php\njson_encode(['amount' => 12], JSON_THROW_ON_ERROR);\n", 0},
		{"<?php other();", 0},
		{"<?php json_encode(-INF);", 1},
		{"<?php json_encode(NAN,JSON_PARTIAL_OUTPUT_ON_ERROR);", 0},
		{"<?php json_encode(INF,$flags);", 0},
		{"<?php json_encode([1,['x'=>-INF]]);", 1},
		{"<?php json_encode([[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[INF]]]]]]]]]]]]]]]]]]]]]]]]]]]]]]]]]]);", 0},
		{"<?php function unreachable(){return;\njson_encode(['amount' => INF], JSON_THROW_ON_ERROR);\n}", 0},
		{"<?php \njson_encode(['amount' => INF], JSON_THROW_ON_ERROR);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeNonFiniteNumber"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/JsonEncodeNonFiniteNumber/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeNonFiniteNumber"}})
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

func BenchmarkContracts(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeNonFiniteNumber"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php json_encode(['data'=>[1,2,3,['value'=>INF]]]);`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}
