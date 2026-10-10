package closurecapturedvaluewrite

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$n=0; ($f=function()use($n){++$n;}); $f(); echo $n;", 0},
		{"$n=0; array_map(function()use($n){++$n;},$xs);", 0},
		{"$n=0; $holder->cb=function()use($n){++$n;};", 0},
		{"$n=0; $f=function()use($n){++$n;};", 0},
		{"$n=0; $f=function()use($n){++$n;}; echo $n;", 0},
		{"$n=0; $f=function()use($n){++$n;}; 1+2;", 0},
		{"$n=0; $f=function()use($n){++$n;}; other(); echo $n;", 0},
		{"$n=0; $f=function()use($n){++$n;}; $f();", 0},
		{"$n=0; $f=function()use($n){++$n;}; $f(); return $n;", 0},
		{"$n=0; $f=function()use($n){$n=4;}; $f(); echo $n;", 1},
		{"$n=0; $f=function()use($n){return ++$n;}; $f(); echo $n;", 0},
		{"$n=0; $f=function()use($n){$g=function(){return 1;};}; $f(); echo $n;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ClosureCapturedValueWrite"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
