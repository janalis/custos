package imagedecodefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A result decoded from a project call or a property is guarded by the facts
// of the paths that reach the use, like any other tracked local.
func TestLocalGuardWalk(t *testing.T) {
	const produce = "$n=imagecreatefromstring(load($b));"
	const decode = "$n=imagecreatefromstring($b);"
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
		{"property bytes unguarded", "$n=imagecreatefromstring($this->d);imagepng($n);", 1},
		{"property bytes or exit", "$n=imagecreatefromstring($this->d);$n!==false or exit;imagepng($n);", 0},
		{"property bytes early exit", "$n=imagecreatefromstring($this->d);if(!$n){return;}imagepng($n);", 0},
		// The forms below are not facts of the shared flow analysis: a disjunction
		// that held or a conjunction that failed with both operands rejecting a
		// failed decode, and assert(). The rule's own walk decides them.
		{"branch of held disjunction", decode + "if($n||$n instanceof GdImage){imagepng($n);}", 0},
		{"branch of held disjunction one side", decode + "if($n||$x){imagepng($n);}", 1},
		{"ternary then of held disjunction", decode + "return ($n||$n!==false)?imagesx($n):0;", 0},
		{"ternary else of failed conjunction", decode + "return (!$n&&$n===false)?0:imagesx($n);", 0},
		{"ternary else of failed conjunction one side", decode + "return (!$n&&$x)?0:imagesx($n);", 1},
		{"and operand of held disjunction", decode + "($n||is_object($n))&&imagepng($n);", 0},
		{"or operand of failed conjunction", decode + "(!$n&&is_bool($n))||imagepng($n);", 0},
		{"else of failed conjunction", decode + "if(!$n&&empty($n)){echo 1;}else{imagepng($n);}", 0},
		{"else after failed elseif conjunction", decode + "if($x){echo 1;}elseif(!$n&&$n==false){echo 2;}else{imagepng($n);}", 0},
		{"exit on conjunction", decode + "if(!$n&&$n===false){return;}imagepng($n);", 0},
		{"assert conjunction", decode + "assert($n&&$x);imagepng($n);", 0},
		{"assert disjunction", decode + "assert($n||$x);imagepng($n);", 1},
		{"assert instanceof", decode + "assert($n instanceof GdImage);imagepng($n);", 0},
		{"assert not instanceof", decode + "assert(!$n instanceof GdImage);imagepng($n);", 1},
		{"assert not empty", decode + "assert(!empty($n));imagepng($n);", 0},
		{"assert empty", decode + "assert(empty($n));imagepng($n);", 1},
		{"assert is_object", decode + "assert(is_object($n));imagepng($n);", 0},
		{"assert is_resource", decode + "assert(is_resource($n));imagepng($n);", 0},
		{"assert not is_object", decode + "assert(!is_object($n));imagepng($n);", 1},
		{"assert not is_bool", decode + "assert(!is_bool($n));imagepng($n);", 0},
		{"assert is_bool", decode + "assert(is_bool($n));imagepng($n);", 1},
		{"assert on earlier value", decode + "assert(is_object($n));$n=imagecreatefromstring($b);imagepng($n);", 1},
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
