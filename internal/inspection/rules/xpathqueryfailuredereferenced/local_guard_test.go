package xpathqueryfailuredereferenced

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts do not combine two operands that each exclude false (a true
// disjunction, a false conjunction), so these guards are decided by the
// rule's own walk.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=$xp->query($q);"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"and operand", produce + "return ($n||$n!==false)&&$n->length;", 0},
		{"ternary then", produce + "return ($n||$n!==false)?$n->length:0;", 0},
		{"ternary else", produce + "return (!$n&&$n===false)?0:$n->length;", 0},
		{"branch body", produce + "if($n!==false||$n instanceof DOMNodeList){return $n->length;}", 0},
		{"branch body one operand only", produce + "if($n||$x){return $n->length;}", 1},
		{"else of failure test", produce + "if(!$n&&$n===false){echo 1;}else{return $n->length;}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(!$n&&$n===false){echo 2;}else{return $n->length;}", 0},
		{"early exit", produce + "if($n===false&&!$n){return 0;}return $n->length;", 0},
		{"early exit one operand only", produce + "if($n===false&&$x){return 0;}return $n->length;", 1},
		{"nested conjunction", produce + "if(($n&&$x)||$n){return $n->length;}", 0},
		{"assignment tested in place", "if(($n=$xp->query($q))||$n instanceof DOMNodeList){return $n->length;}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"XPathQueryFailureDereferenced"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f(DOMXPath $xp,$q,$x){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
