package pregreplacefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts prove nothing about a result computed from a project call,
// so each guard below is decided by the rule's own walk.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=preg_replace($t,$r,load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "strlen($n);", 1},
		{"branch body", produce + "if($n!==null){strlen($n);}", 0},
		{"elseif own condition", produce + "if($x){echo 1;}elseif(is_string($n)){strlen($n);}", 0},
		{"ternary else", produce + "return $n===null?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n===null?strlen($n):0;", 1},
		{"and operand", produce + "return !is_null($n)&&strlen($n);", 0},
		{"isset", produce + "if(isset($n)){strlen($n);}", 0},
		{"truthiness is not a check", produce + "if($n){strlen($n);}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PregReplaceFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f($t,$r,$b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
