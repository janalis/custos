package tempnamdirectoryfallbackunchecked

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewTempnamDirectoryFallbackUnchecked(t *testing.T) {
	const make = "function f($dir,$s){$p=tempnam($dir,'job');if($p===false){return;}"
	const use = "file_put_contents($p,$s);}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"unchecked", make + use, 1},
		{"top level", "$p=tempnam($dir,'job');if($p===false){exit;}if(realpath(dirname($p))!==realpath($dir)){exit;}file_put_contents($p,$s);", 0},
		{"path rebound after guard", make + "$q=$p;if(realpath(dirname($p))!==realpath($dir)){return;}$p=$q;" + use, 1},
		{"saved comparison", make + "$same=realpath(dirname($p))===realpath($dir);if(!$same){return;}" + use, 0},
		{"mismatch exit with else", make + "if(realpath(dirname($p))!==realpath($dir)){return;}else{$s.='x';}" + use, 0},
		{"match as short circuit", make + "realpath(dirname($p))===realpath($dir)&&file_put_contents($p,$s);}", 0},
		{"mismatch compared with false", make + "if((realpath(dirname($p))!==realpath($dir))===true){return;}" + use, 0},
		{"jump label after guard", make + "if(realpath(dirname($p))!==realpath($dir)){return;}back:file_put_contents($p,$s);goto back;}", 1},
		{"path no longer a temporary file", make + "if(realpath(dirname($p))!==realpath($dir)){return;}$p=$s;" + use, 0},
		{"mismatch branch", make + "if(realpath(dirname($p))!==realpath($dir)){return;}" + use, 0},
		{"swapped operands", make + "if(realpath($dir)!==realpath(dirname($p))){unlink($p);return;}" + use, 0},
		{"loose mismatch", make + "if(realpath(dirname($p))!=realpath($dir)){return;}" + use, 0},
		{"parenthesized operands", make + "if((realpath((dirname(($p)))))!==(realpath($dir))){return;}" + use, 0},
		{"mismatch disjunct", "function f($dir,$s){$p=tempnam($dir,'job');if($p===false||realpath($dir)!==realpath(dirname($p))){return;}" + use, 0},
		{"mismatch conjunct proves nothing", make + "if($s&&realpath(dirname($p))!==realpath($dir)){return;}" + use, 1},
		{"mismatch branch continues", make + "if(realpath(dirname($p))!==realpath($dir)){log_it();}" + use, 1},
		{"mismatch in outer block", make + "if(realpath(dirname($p))!==realpath($dir)){return;}if($s){" + use + "}", 0},
		{"match branch", make + "if(realpath(dirname($p))===realpath($dir)){" + use + "}", 0},
		{"match branch swapped loose", make + "if(realpath($dir)==realpath(dirname($p))){" + use + "}", 0},
		{"match conjunct", "function f($dir,$s){$p=tempnam($dir,'job');if($p!==false&&realpath(dirname($p))===realpath($dir)){" + use + "}", 0},
		{"match disjunct proves nothing", make + "if($s||realpath(dirname($p))===realpath($dir)){" + use + "}", 1},
		{"match else branch", make + "if(realpath(dirname($p))===realpath($dir)){}else{" + use + "}", 1},
		{"match used as early exit", make + "if(realpath(dirname($p))===realpath($dir)){return;}" + use, 1},
		{"mismatch used as enclosing branch", make + "if(realpath(dirname($p))!==realpath($dir)){" + use + "}", 1},
		{"uncanonical comparison", make + "if(dirname($p)!==$dir){return;}" + use, 1},
		{"uncanonical required directory", make + "if(realpath(dirname($p))!==$dir){return;}" + use, 1},
		{"other variable", make + "if(realpath(dirname($q))!==realpath($dir)){return;}" + use, 1},
		{"parent of another function", make + "if(realpath(basename($p))!==realpath($dir)){return;}" + use, 1},
		{"other operator", make + "if(realpath(dirname($p))<realpath($dir)){return;}" + use, 1},
		{"non comparison condition", make + "if(is_dir($dir)){" + use + "}", 1},
		{"guard precedes new file", make + "if(realpath(dirname($p))!==realpath($dir)){return;}$p=tempnam($dir,'k');" + use, 1},
		{"match guard precedes new file", make + "if(realpath(dirname($p))===realpath($dir)){$p=tempnam($dir,'k');" + use + "}", 1},
		{"guard outside closure", make + "if(realpath(dirname($p))!==realpath($dir)){return;}$g=function()use($dir,$s){$p=tempnam($dir,'k');" + use + ";}", 1},
		{"system directory", "function f($s){$p=tempnam(sys_get_temp_dir(),'job');" + use, 0},
		{"system directory parenthesized", "function f($s){$p=tempnam((sys_get_temp_dir()),'job');" + use, 0},
		{"system directory variable", "function f($s){$tmp=sys_get_temp_dir();$p=tempnam($tmp,'job');" + use, 0},
		{"system directory named argument", "function f($s){$p=tempnam(prefix:'job',directory:sys_get_temp_dir());" + use, 0},
		{"system directory subfolder", "function f($s){$p=tempnam(sys_get_temp_dir().'/app','job');" + use, 1},
		{"shadowed system directory", "namespace App;function sys_get_temp_dir(){return '/private';}function f($s){$p=tempnam(sys_get_temp_dir(),'job');" + use, 1},
		{"read handle", make + "$h=fopen($p,'r');}", 1},
		{"inline temporary file", "file_put_contents(tempnam($dir,'p'),'x');", 1},
		{"other path", "function f($s){$p='/tmp/known';" + use, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"TempnamDirectoryFallbackUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
