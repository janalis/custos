package jsonencodeknownrecursivevalue

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
		{"<?php\n$a = []; $a['loop'] = &$a; json_encode($a, JSON_THROW_ON_ERROR);\n", 1},
		{"<?php\n$a = ['loop' => []]; json_encode($a, JSON_THROW_ON_ERROR);\n", 0},
		{"<?php other();", 0},
		{"<?php $a=[];$b=[];$a['b']=&$b;$b['a']=&$a;json_encode($a);", 1},
		{"<?php $a=[];$a['x']=&$a;json_encode($a,JSON_PARTIAL_OUTPUT_ON_ERROR);", 0},
		{"<?php $a=[];$a['x']=&$a;json_encode($a,$flags);", 0},
		{"<?php json_encode([]);", 0},
		{"<?php $a=[];if($ok){$a=[];}$a['x']=&$a;json_encode($a);", 0},
		{"<?php $a=[];$a[$key]=&$a;json_encode($a);", 0},
		{"<?php $a=[];$a['x']=&$a;unset($a['x']);json_encode($a);", 0},
		{"<?php $a=[];$a['x']=&$a;change($a);json_encode($a);", 0},
		{"<?php $a=[];$a['x']=&$a;$o->change();json_encode($a);", 0},
		{"<?php $a=[];$a['x']=&$a;$a['x']=[];json_encode($a);", 0},
		{"<?php $a=[];$a['x']=&$a;$a=[];json_encode($a);", 0},
		{"<?php $a=[];$b=[];$a['b']=&$b;json_encode($a);", 0},
		{"<?php $a=[];$b=[];$c=[];$a['b']=&$b;$a['c']=&$c;$c['b']=&$b;json_encode($a);", 0},
		{"<?php function unreachable(){return;\n$a = []; $a['loop'] = &$a; json_encode($a, JSON_THROW_ON_ERROR);\n}", 0},
		{"<?php \n$a = []; $a['loop'] = &$a; json_encode($a, JSON_THROW_ON_ERROR);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeKnownRecursiveValue"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/JsonEncodeKnownRecursiveValue/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeKnownRecursiveValue"}})
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeKnownRecursiveValue"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php $a=[];$b=[];$a['b']=&$b;$b['a']=&$a;json_encode($a);`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}
