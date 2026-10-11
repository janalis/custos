package imagedecodefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The flow facts prove nothing about a result computed from a project call,
// so each guard below is decided by the rule's own walk.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=imagecreatefromstring(load($b));"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"unguarded", produce + "imagepng($n);", 1},
		{"ternary then", produce + "return $n?imagesx($n):0;", 0},
		{"ternary else", produce + "return !$n?0:imagesx($n);", 0},
		{"ternary wrong arm", produce + "return $n?0:imagesx($n);", 1},
		{"and operand", produce + "$n&&imagepng($n);", 0},
		{"else of failure test", produce + "if(!$n){echo 1;}else{imagepng($n);}", 0},
		{"else after failed elseif", produce + "if($x){echo 1;}elseif(!$n){echo 2;}else{imagepng($n);}", 0},
		{"unrelated chain before use", produce + "if($x){echo 1;}elseif($l){echo 2;}imagepng($n);", 1},
		{"conjunction", produce + "if($n&&$x){imagepng($n);}", 0},
		{"instanceof", produce + "if($n instanceof GdImage){imagepng($n);}", 0},
		{"is_object", produce + "if(is_object($n)){imagepng($n);}", 0},
		{"is_bool exit", produce + "if(is_bool($n)){return;}imagepng($n);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ImageDecodeFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("guard.php", []byte("<?php function f($b,$x,$l){"+tc.source+"}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("invalid source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
