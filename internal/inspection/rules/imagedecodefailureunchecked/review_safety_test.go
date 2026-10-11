package imagedecodefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewGuards(t *testing.T) {
	const decode = "$im=imagecreatefromstring($b);"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"negated early return", decode + "if(!$im){return;}imagepng($im);", 0},
		{"truthy branch", decode + "if($im){imagepng($im);}", 0},
		{"alternative syntax", decode + "if(false!==$im):imagepng($im);endif;", 0},
		{"else of failure test", decode + "if($im===false){echo 1;}else{imagepng($im);}", 0},
		{"elseif own condition", decode + "if($c){echo 1;}elseif($im){imagepng($im);}", 0},
		{"elseif after failed test", decode + "if(!$im){echo 1;}elseif($c){imagepng($im);}", 0},
		{"else after failed elseif", decode + "if($c){echo 1;}elseif(!$im){echo 2;}else{imagepng($im);}", 0},
		{"later elseif after failed elseif", decode + "if($c){echo 1;}elseif(!$im){echo 2;}elseif($d){imagepng($im);}", 0},
		{"else with unrelated chain", decode + "if($c){echo 1;}elseif($d){echo 2;}else{imagepng($im);}", 1},
		{"elseif unrelated", decode + "if($c){echo 1;}elseif($d){imagepng($im);}", 1},
		{"use in if condition", decode + "if(imagesx($im)>1){echo 1;}", 1},
		{"use in elseif condition", decode + "if($c){echo 1;}elseif(imagesx($im)>1){echo 2;}", 1},
		{"use in elseif condition after test", decode + "if(!$im){echo 1;}elseif(imagesx($im)>1){echo 2;}", 0},
		{"failure branch", decode + "if(!$im){imagepng($im);}", 1},
		{"ternary else", decode + "return $im===false?0:imagesx($im);", 0},
		{"ternary then", decode + "return $im?imagesx($im):0;", 0},
		{"ternary wrong arm", decode + "return $im?0:imagesx($im);", 1},
		{"ternary unrelated", decode + "return $c?imagesx($im):imagesy($im);", 2},
		{"short ternary", decode + "return $c?:imagesx($im);", 1},
		{"and operand", decode + "$im!==false&&imagepng($im);", 0},
		{"and keyword operand", decode + "$im and imagepng($im);", 0},
		{"or operand", decode + "$im===false||imagepng($im);", 0},
		{"or operand unguarded", decode + "$im||imagepng($im);", 1},
		{"left operand", decode + "imagepng($im)&&$im;", 1},
		{"arithmetic operand", decode + "echo 1+imagesx($im);", 1},
		{"instanceof", decode + "if(!$im instanceof GdImage){throw new E();}imagepng($im);", 0},
		{"parenthesised instanceof", decode + "if(!($im instanceof \\GdImage)){return;}imagepng($im);", 0},
		{"inline assignment", "if(($im=imagecreatefromstring($b))===false){throw new E();}imagepng($im);", 0},
		{"negated inline assignment", "if(!$im=imagecreatefromstring($b)){return;}imagepng($im);", 0},
		{"inline assignment of other variable", decode + "if(!$x=imagecreatefromstring($b)){return;}imagepng($im);", 1},
		{"inline assignment of other call", decode + "if(!$im2=other($b)){return;}imagepng($im);", 1},
		{"or die", "$im=imagecreatefromstring($b) or die('x');imagepng($im);", 0},
		{"or throw", decode + "$im||throw new E();imagepng($im);", 0},
		{"and exit", decode + "!$im&&exit(1);imagepng($im);", 0},
		{"and exit wrong sense", decode + "$im&&exit(1);imagepng($im);", 1},
		{"or call", decode + "$im||fail();imagepng($im);", 1},
		{"sum statement", decode + "$im+1;imagepng($im);", 1},
		{"assert", decode + "assert($im!==false);imagepng($im);", 0},
		{"assert unrelated", decode + "assert($c);imagepng($im);", 1},
		{"assert without argument", decode + "assert();imagepng($im);", 1},
		{"other call statement", decode + "check($im!==false);imagepng($im);", 1},
		{"is_bool", decode + "if(is_bool($im)){return;}imagepng($im);", 0},
		{"is_object", decode + "if(is_object($im)){imagepng($im);}", 0},
		{"is_resource", decode + "if(!is_resource($im)){return;}imagepng($im);", 0},
		{"is_string", decode + "if(is_string($im)){return;}imagepng($im);", 1},
		{"empty", decode + "if(empty($im)){return;}imagepng($im);", 0},
		{"loose false", decode + "if($im==false){return;}imagepng($im);", 0},
		{"loose not false", decode + "if($im!=false){imagepng($im);}", 0},
		{"compared with true", decode + "if($im===true){return;}imagepng($im);", 1},
		{"compared with null", decode + "if($im===null){return;}imagepng($im);", 1},
		{"ordering comparison", decode + "if($im<1){return;}imagepng($im);", 1},
		{"other unary", decode + "if(-$im){imagepng($im);}", 1},
		{"or of failure tests", decode + "if(null===$im||false===$im){return;}imagepng($im);", 0},
		{"or with later use", decode + "if($im===false||imagesx($im)<1){return;}imagepng($im);", 0},
		{"and needs both", decode + "if(!$im&&$c){return;}imagepng($im);", 1},
		{"and success", decode + "if($im&&$c){imagepng($im);}", 0},
		{"or success needs both", decode + "if($im||$c){imagepng($im);}", 1},
		{"terminating else", decode + "if($im!==false){}else{return;}imagepng($im);", 0},
		{"terminating else wrong sense", decode + "if(!$im){}else{return;}imagepng($im);", 1},
		{"elseif guard statement", decode + "if(!$im){return;}elseif($c){return;}imagepng($im);", 0},
		{"non-terminating guard", decode + "if(!$im){fail();}imagepng($im);", 1},
		{"continue", decode + "foreach($l as $_){if(!$im){continue;}imagepng($im);}", 0},
		{"guard after use in loop", decode + "foreach($l as $_){imagepng($im);if(!$im){return;}}", 1},
		{"nested use after guard", decode + "if(!$im){return;}if($c){echo imagesx($im);}", 0},
		{"guard on earlier value", decode + "if(!$im){return;}$im=imagecreatefromstring($b);imagepng($im);", 1},
		{"guard on other variable", decode + "if(!$other){return;}imagepng($im);", 1},
		{"guard outside closure", decode + "if(!$im){return;}$f=function()use($b){$im=imagecreatefromstring($b);imagepng($im);};", 1},
		{"unrelated statements before", "echo 1;" + decode + "echo 2;while($c){}imagepng($im);", 1},
		{"incomplete ternary", decode + "return $c?imagesx($im):", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ImageDecodeFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php function f($b,$c,$d,$l){"+tc.source+"}"), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
