package xsltransformationfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Each guard below is read twice. In a scope the flow follows, its facts
// prove it although the result is computed from a project call. A scope with
// `goto` is not followed: the producer is still found lexically and the
// rule's own walk decides.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=$p->transformToXML(load($d));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "echo $n;", 1},
		{"branch body", produce + "if($n){echo $n;}", 0},
		{"else of failure test", produce + "if(!$n){echo 1;}else{echo $n;}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(empty($n)){echo 2;}else{echo $n;}", 0},
		{"else of unrelated chain", produce + "if($x){echo 1;}elseif($file){echo 2;}else{echo $n;}", 1},
		{"elseif own condition", produce + "if($x){echo 1;}elseif(is_string($n)){echo $n;}", 0},
		{"ternary then", produce + "return $n?file_put_contents($file,$n):0;", 0},
		{"ternary else", produce + "return !$n?0:file_put_contents($file,$n);", 0},
		{"ternary wrong arm", produce + "return $n?0:file_put_contents($file,$n);", 1},
		{"and operand", produce + "return $n&&file_put_contents($file,$n);", 0},
		{"early exit on loose false", produce + "if($n==false){return;}echo $n;", 0},
		{"early exit on unrelated test", produce + "if($x){return;}echo $n;", 1},
		{"early exit on yoda loose false", produce + "if(false==$n){return;}echo $n;", 0},
		{"early exit on empty", produce + "if(empty($n)){return;}echo $n;", 0},
		{"early exit before other statements", produce + "if(!$n){return;}echo 1;echo $n;", 0},
		{"loose branch", produce + "if($n!=false){echo $n;}", 0},
		{"conjunction branch", produce + "if($x&&$n){echo $n;}", 0},
		{"disjunction branch", produce + "if($x||$n){echo $n;}", 1},
		{"disjunction exit", produce + "if($x||!$n){return;}echo $n;", 0},
		{"or operand", produce + "return !$n||file_put_contents($file,$n);", 0},
		{"or operand wrong sense", produce + "return $n||file_put_contents($file,$n);", 1},
		{"compared to true", produce + "if($n!=true){echo $n;}", 1},
		{"disjunction of two rejecting tests", produce + "if($n||is_string($n)){echo $n;}", 0},
		{"disjunction with one rejecting test", produce + "if($x||is_string($n)){echo $n;}", 1},
		{"disjunction with a failure test", produce + "if($n||$n===false){echo $n;}", 1},
		{"exit on conjunction of two failure tests", produce + "if($n===false&&!is_string($n)){return;}echo $n;", 0},
		{"exit on conjunction with one failure test", produce + "if($n===false&&$x){return;}echo $n;", 1},
		{"exit on conjunction with a passing test", produce + "if(!$n&&is_string($n)){return;}echo $n;", 1},
		{"else of conjunction of two failure tests", produce + "if(!$n&&$n==false){echo 1;}else{echo $n;}", 0},
		{"ternary on disjunction of two rejecting tests", produce + "return ($n!==false||is_string($n))?file_put_contents($file,$n):0;", 0},
	} {
		for _, tail := range []string{"", "goto done;done:;"} {
			t.Run(tc.name+" "+tail, func(t *testing.T) {
				e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"XslTransformationFailureConsumed"}})
				if err != nil {
					t.Fatal(err)
				}
				f := syntax.Parse("guard.php", []byte("<?php function f(XSLTProcessor $p,$d,$x,$file){"+tc.source+tail+"}"), syntax.Options{})
				if len(f.Errors) > 0 {
					t.Fatalf("invalid source: %+v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.want {
					t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
				}
			})
		}
	}
}
