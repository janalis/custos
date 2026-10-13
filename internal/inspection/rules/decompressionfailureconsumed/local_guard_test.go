package decompressionfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A result computed from a project call or a property is guarded by the facts
// of the paths that reach the use, like any other tracked local.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=gzdecode(load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "strlen($n);", 1},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif($n===false){echo 2;}else{strlen($n);}", 0},
		{"else of unrelated chain", produce + "if($x){echo 1;}elseif($b){echo 2;}else{strlen($n);}", 1},
		{"early exit", produce + "if($n===false){return;}strlen($n);", 0},
		// The shared flow analysis reads a truthiness test as excluding false.
		{"early exit on truthiness", produce + "if(!$n){return;}strlen($n);", 0},
		{"ternary arm", produce + "return $n===false?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n===false?strlen($n):0;", 1},
		{"and operand", produce + "return $n!==false&&strlen($n);", 0},
		{"while body", produce + "while($n!==false){strlen($n);break;}", 0},
		{"reassigned after guard", produce + "if($n===false){return;}$n=gzdecode($x);strlen($n);", 1},
		{"property input unguarded", "$n=gzdecode($this->d);strlen($n);", 1},
		{"property input or exit", "$n=gzdecode($this->d);$n!==false or exit;strlen($n);", 0},
		{"property input early exit", "$n=gzdecode($this->d);if($n===false){return;}strlen($n);", 0},
		{"guard reached by another entry", produce + "switch($x){case 1:if($n===false){return;}case 2:strlen($n);}", 1},
		{"finally after guarded return", produce + "try{if($n===false){return;}}finally{strlen($n);}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DecompressionFailureConsumed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f($b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
