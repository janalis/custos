package arrayrandinvalidcount

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
		{"<?php\narray_rand(['oak', 'elm'], 4);\n", 1},
		{"<?php\narray_rand(['oak', 'elm'], 1);\n", 0},
		{"<?php other();", 0},
		{"<?php array_rand([]);", 1},
		{"<?php array_rand($items,0);", 1},
		{"<?php array_rand($items,$n);", 0},
		{"<?php array_rand([1],$n);", 0},
		{"<?php array_rand($items);", 0},
		{"<?php $v = []; $v['x']++; array_rand($v);", 0},
		{"<?php $n = 0; ++$n; array_rand($items, ($n));", 0},
		{"<?php function f($xs) { $v = []; foreach ($xs as $x) { @$v[$x]++; } return array_rand($v); }", 0},
		{"<?php function f($xs, $a) { $n = 0; foreach ($xs as $x) { if ($x) { $n--; } } return array_rand($a, $n); }", 0},
		{"<?php function f() { $v = []; foreach ([2] as $v[9]) {} return array_rand($v); }", 0},
		{"<?php function f() { $v = []; foreach ([2 => 1] as $v['k'] => $x) {} return array_rand($v); }", 0},
		{"<?php function f() { $v = []; $g = function () use (&$v) { $v[] = 1; }; $g(); return array_rand($v); }", 0},
		{"<?php function f() { $v = []; $g = function () use ($v) {}; return array_rand($v); }", 1},
		{"<?php function f() { $v = []; $box = [&$v]; $box[0][] = 2; return array_rand($v); }", 0},
		{"<?php function f() { $v = []; $box = [$v]; return array_rand($v); }", 1},
		{"<?php function f($o) { $v = []; EXTRACT($o); return array_rand($v); }", 0},
		{"<?php function f($xs) { $v = []; $i = 0; foreach ($xs as $x) { $i++; other($x); } return array_rand($v); }", 1},
		{"<?php function f($a) { $n = 0; $$a = 1; return array_rand($a, $n); }", 0},
		{"<?php function f($a, $i) { $n = 0; $i++; return array_rand($a, $n); }", 1},
		{"<?php function f($xs) { $v = []; foreach ($xs as $x) { $v[$x]++; } $w = $v; return array_rand($w); }", 0},
		{"<?php function f($xs, $a) { $n = 0; foreach ($xs as $x) { $n++; } $m = $n; $k = $m; return array_rand($a, $k); }", 0},
		{"<?php function f($a) { $n = 0; $m = $n; $n = $m; return array_rand($a, $n); }", 1},
		{"<?php function f() { $v = []; $w = $v; return array_rand($w); }", 1},
		{"<?php function f($a) { $n = 0; foreach ([&$n] as &$slot) { $slot = 1; } return array_rand($a, $n); }", 0},
		{"<?php function unreachable(){return;\narray_rand(['oak', 'elm'], 4);\n}", 0},
		{"<?php \narray_rand(['oak', 'elm'], 4);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayRandInvalidCount"}})
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
		marked, err := os.ReadFile("../../../../testdata/rules/ArrayRandInvalidCount/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayRandInvalidCount"}})
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
