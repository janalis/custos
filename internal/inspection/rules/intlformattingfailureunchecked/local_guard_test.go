package intlformattingfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts prove nothing about a result computed from a project call,
// so each guard below is decided by the rule's own walk.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=$fm->format(load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "strlen($n);", 1},
		{"ternary then", produce + "return $n?strlen($n):0;", 0},
		{"ternary else", produce + "return !$n?0:strlen($n);", 0},
		{"ternary wrong arm", produce + "return $n?0:strlen($n);", 1},
		{"and operand", produce + "return $n&&strlen($n);", 0},
		{"else of failure test", produce + "if(!$n){echo 1;}else{strlen($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(!$n){echo 2;}else{strlen($n);}", 0},
		{"empty exit", produce + "if(empty($n)){return;}strlen($n);", 0},
		{"is_bool exit", produce + "if(is_bool($n)){return;}strlen($n);", 0},
		{"is_scalar", produce + "if(is_scalar($n)){strlen($n);}", 1},
		{"assignment tested in place", "if($n=$fm->format(load($x))){strlen($n);}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IntlFormattingFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f(NumberFormatter $fm,$b,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
