package stringrepeatnegativecount

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nstr_repeat('cd', -3);\n", 1},
		{"<?php\nstr_repeat('cd', 3);\n", 0},
		{"<?php other();", 0},
		{"<?php str_repeat(\"x\",$n);", 0},
		{"<?php str_repeat(times:-2,string:\"x\");", 1},
		{"<?php $n = -2; str_repeat('x', $n);", 1},
		{"<?php function f($o) { extract($o); $n = -2; return str_repeat('x', $n); }", 1},
		{"<?php function f() { $n = -5; $n++; $n = -2; return str_repeat('x', $n); }", 1},
		{"<?php function f() { $n = -5; $n--; $n = -2; return str_repeat('x', $n); }", 1},
		{"<?php function f() { $n = -2; $g = function () use (&$n) { $n = 3; }; return str_repeat('x', $n) . gettype($g); }", 1},
		{"<?php function f($xs, $o) { $r = ''; foreach ($xs as $x) { $n = -2; $r = str_repeat('x', $n); extract($o); } return $r; }", 1},
		{"<?php function unreachable(){return;\nstr_repeat('cd', -3);\n}", 0},
		{"<?php \nstr_repeat('cd', -3);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"StringRepeatNegativeCount"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
