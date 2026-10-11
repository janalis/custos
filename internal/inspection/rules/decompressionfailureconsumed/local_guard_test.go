package decompressionfailureconsumed

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts prove nothing about a result computed from a project call,
// so each guard below is decided by the rule's own walk.
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
		{"early exit on truthiness", produce + "if(!$n){return;}strlen($n);", 1},
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
