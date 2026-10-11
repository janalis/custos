package xsltransformationfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts prove nothing about a result computed from a project call,
// so each guard below is decided by the rule's own walk.
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"XslTransformationFailureConsumed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f(XSLTProcessor $p,$d,$x,$file){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
