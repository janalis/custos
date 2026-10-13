package unicodenormalizationfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A result normalized from a project call or a property is guarded by the
// facts of the paths that reach the use, like any other tracked local.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=normalizer_normalize(load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "strlen($n);", 1},
		{"ternary then", produce + "return $n!==false?strlen($n):0;", 0},
		{"ternary else", produce + "return $n===false?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n===false?strlen($n):0;", 1},
		{"and operand", produce + "return $n!==false&&strlen($n);", 0},
		{"elseif own condition", produce + "if($x){echo 1;}elseif(is_string($n)){strlen($n);}", 0},
		{"else of failure test", produce + "if(!is_string($n)){echo 1;}else{strlen($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif($n===false){echo 2;}else{strlen($n);}", 0},
		{"conjunction", produce + "if($x&&$n!==false){strlen($n);}", 0},
		{"is_string without argument", produce + "if(is_string()){strlen($n);}", 1},
		{"early exit on truthiness", produce + "if(!$n){return;}strlen($n);", 0},
		{"property text unguarded", "$n=normalizer_normalize($this->d);strlen($n);", 1},
		{"property text or exit", "$n=normalizer_normalize($this->d);$n!==false or exit;strlen($n);", 0},
		{"property text early exit", "$n=normalizer_normalize($this->d);if($n===false){return;}strlen($n);", 0},
		// An is_string() test compared with a boolean literal is not a fact of the
		// shared flow analysis: the rule's own walk decides the forms below.
		{"ternary then of compared predicate", produce + "return is_string($n)===true?strlen($n):0;", 0},
		{"ternary else of compared predicate", produce + "return is_string($n)===false?0:strlen($n);", 0},
		{"ternary wrong arm of compared predicate", produce + "return is_string($n)===true?0:strlen($n);", 1},
		{"and operand of compared predicate", produce + "return true===is_string($n)&&strlen($n);", 0},
		{"or operand of compared predicate", produce + "return is_string($n)!==true||strlen($n);", 0},
		{"or operand wrong sense", produce + "return is_string($n)===true||strlen($n);", 1},
		{"branch of compared predicate", produce + "if(is_string($n)===true){strlen($n);}", 0},
		{"branch of other compared predicate", produce + "if(is_string($x)===true){strlen($n);}", 1},
		{"elseif of compared predicate", produce + "if($x){echo 1;}elseif(is_string($n)!==false){strlen($n);}", 0},
		{"else of compared predicate", produce + "if(is_string($n)!==true){echo 1;}else{strlen($n);}", 0},
		{"else after failed compared elseif", produce + "if($x){echo 1;}elseif(is_string($n)===false){echo 2;}else{strlen($n);}", 0},
		{"negated compared predicate", produce + "if(!(is_string($n)===false)){strlen($n);}", 0},
		{"conjunction with compared predicate", produce + "if($x&&is_string($n)===true){strlen($n);}", 0},
		{"disjunction with compared predicate", produce + "if($x||is_string($n)===true){strlen($n);}", 1},
		{"exit on compared predicate", produce + "if(is_string($n)===false){return;}strlen($n);", 0},
		{"compared predicate on earlier value", produce + "if(is_string($n)===false){return;}$n=normalizer_normalize($b);strlen($n);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnicodeNormalizationFailureConsumed"}})
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
