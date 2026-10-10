package parseurlfailuredereferenced

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
		{"<?php\n$parts = parse_url($raw); echo $parts['host'];\n", 1},
		{"<?php\n$parts = parse_url($raw); if ($parts === false) { throw new RuntimeException(); } echo $parts['host'];\n", 0},
		{"<?php other();", 0},
		{"<?php echo $x['a'];", 0},
		{"<?php echo parse_url($s,PHP_URL_HOST)[0];", 0},
		{"<?php echo parse_url($s,-1)['host'];", 1},
		{"<?php echo parse_url($s,$part)['host'];", 0},
		{"<?php namespace N;function parse_url($s){return [];}echo parse_url($s)['host'];", 0},
		{"<?php function unreachable(){return;\n$parts = parse_url($raw); echo $parts['host'];\n}", 0},
		{"<?php \n$parts = parse_url($raw); echo $parts['host'];\n function broken(", 0},
		{"<?php $a=parse_url($p);$alias=&$a;mutate($alias);echo $a['host'];", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ParseUrlFailureDereferenced"}})
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
	for _, name := range []string{"basic", "negative", "audit-negative-2", "literal-urls"} {
		marked, err := os.ReadFile("../../../../testdata/rules/ParseUrlFailureDereferenced/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ParseUrlFailureDereferenced"}})
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

func BenchmarkLiteralURLRead(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ParseUrlFailureDereferenced"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php echo parse_url('https://example.test:443/path')['host'];"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(file)
	}
}
