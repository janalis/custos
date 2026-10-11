package exhaustedgeneratorreused

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewIdentityAndRebinding pins the review findings: separate generator
// calls are separate identities, and a variable that code outside the local
// flow can rebind carries no exhaustion proof.
func TestReviewIdentityAndRebinding(t *testing.T) {
	const gen = "function gen(){yield 1;yield 2;} "
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"separate calls", "foreach(gen() as $_){} foreach(gen() as $v){}", 0},
		{"separate calls in function", "function f(){foreach(gen() as $_){} foreach(gen() as $v){}}", 0},
		{"variable variable", "$n='g'; $$n=gen(); foreach($$n as $_){} foreach($$n as $v){}", 0},
		{"same variable", "function f(){$g=gen(); foreach($g as $_){} foreach($g as $v){}}", 1},
		{"no preceding statement", "function f($c){$g=gen(); if($c){foreach($g as $v){}}}", 0},
		{"different variables", "function f(){$g=gen(); $h=gen(); foreach($h as $_){} foreach($g as $v){}}", 0},
		{"first loop by reference", "function f(){$g=gen(); foreach($g as &$r){} foreach($g as $v){}}", 0},
		{"first loop overwrites value", "function f(){$g=gen(); foreach($g as $g){} foreach($g as $v){}}", 0},
		{"first loop overwrites key", "function f(){$g=gen(); foreach($g as $g => $_){} foreach($g as $v){}}", 0},
		{"by-reference capture", "function f(){$g=gen(); $r=function() use (&$g){$g=gen();}; iterator_to_array($g); $r(); foreach($g as $v){}}", 0},
		{"by-reference capture of another variable", "function f(){$g=gen(); $o=1; $r=function() use (&$o){}; iterator_to_array($g); $r(); foreach($g as $v){}}", 1},
		{"by-value capture", "function f(){$g=gen(); $r=function() use ($g){}; iterator_to_array($g); $r(); foreach($g as $v){}}", 1},
		{"global in scope", "function f(){global $g; $g=gen(); iterator_to_array($g); renew(); foreach($g as $v){}}", 0},
		{"global of another variable", "function f(){global $o; $g=gen(); iterator_to_array($g); foreach($g as $v){}}", 1},
		{"global in nested closure", "function f(){$g=gen(); $r=function(){global $g;}; iterator_to_array($g); foreach($g as $v){}}", 1},
		{"top-level variable declared global elsewhere", "function renew(){global $g; $g=gen();} $g=gen(); iterator_to_array($g); renew(); foreach($g as $v){}", 0},
		{"top-level variable", "function other(){global $o;} $g=gen(); iterator_to_array($g); foreach($g as $v){}", 1},
		{"static in scope", "function f(){static $g; $g=gen(); iterator_to_array($g); f(); foreach($g as $v){}}", 0},
		{"static of another variable", "function f(){static $o; $g=gen(); iterator_to_array($g); foreach($g as $v){}}", 1},
		{"static in nested closure", "function f(){$g=gen(); $r=function(){static $g;}; iterator_to_array($g); foreach($g as $v){}}", 1},
		{"extract", "function f(){$g=gen(); iterator_to_array($g); extract(['g'=>gen()]); foreach($g as $v){}}", 0},
		{"parse_str", "function f(){$g=gen(); iterator_to_array($g); parse_str('g=1'); foreach($g as $v){}}", 0},
		{"extract in nested closure", "function f(){$g=gen(); $r=function(){extract([]);}; iterator_to_array($g); foreach($g as $v){}}", 1},
		{"include", "function f(){$g=gen(); iterator_to_array($g); include 'renew.php'; foreach($g as $v){}}", 0},
		{"include in nested closure", "function f(){$g=gen(); $r=function(){include 'x.php';}; iterator_to_array($g); foreach($g as $v){}}", 1},
		{"second iterator_to_array", "function f(){$g=gen(); iterator_to_array($g); return iterator_to_array($g);}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ExhaustedGeneratorReused"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+gen+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
