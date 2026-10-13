package transliterationfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A result computed from a project call is not completely known, yet the
// not-false facts of its paths still hold: the flow proves each guard below.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=transliterator_transliterate($t,load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "strlen($n);", 1},
		{"else of failure test", produce + "if($n===false){echo 1;}else{strlen($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif($n===false){echo 2;}else{strlen($n);}", 0},
		{"elseif own condition", produce + "if($x){echo 1;}elseif(is_string($n)){strlen($n);}", 0},
		{"and operand", produce + "return $n!==false&&strlen($n);", 0},
		{"early exit", produce + "if(!is_string($n)){return;}strlen($n);", 0},
		{"conjunction", produce + "if($x&&$n!==false){strlen($n);}", 0},
		{"disjunction exit", produce + "if($x||$n===false){return;}strlen($n);", 0},
		{"truthiness", produce + "if($n){strlen($n);}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"TransliterationFailureConsumed"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f($t,$b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
