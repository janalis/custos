package streamuseafterclose

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewStreamUseAfterCloseProof(t *testing.T) {
	const open = "$h=fopen($p,'rb');if($h===false){return;}"
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"closed", "function f($p){" + open + "fclose($h);fread($h,1);}", 1},
		{"file scope", "$h=fopen('a','rb');if($h!==false){fclose($h);fread($h,1);}", 1},
		{"suppressed close", "function f($p){" + open + "@fclose($h);fread($h,1);}", 1},
		{"success branch", "function f($p){" + open + "if(fclose($h)){fread($h,1);}}", 1},
		{"success after guard", "function f($p){" + open + "if(!fclose($h)){return;}fread($h,1);}", 1},
		{"stored success", "function f($p){" + open + "$ok=fclose($h);if($ok){fread($h,1);}}", 1},
		{"stored result overwritten", "function f($p,$q){" + open + "$ok=fclose($h);$ok=$q;if(!$ok){fread($h,1);}}", 1},
		{"stored result read in nested scope", "function f($p){" + open + "$ok=fclose($h);$g=function()use($ok){return !$ok;};fread($h,1);}", 1},
		{"short-circuit success", "function f($p){" + open + "fclose($h) and fread($h,1);}", 1},
		{"captured by value", "function f($p){" + open + "$g=function()use($h){};fclose($h);$g();fread($h,1);}", 1},
		{"other variable captured", "function f($p){" + open + "$o=1;$g=function()use(&$o){};fclose($h);$g();fread($h,1);}", 1},
		{"capture in nested scope only", "function f($p){" + open + "$g=function(){$i=function()use(&$h){};};fclose($h);fread($h,1);}", 1},
		{"other call before close", "function f($p){" + open + "strlen($p);fclose($h);fread($h,1);}", 1},

		{"failure branch", "function f($p){" + open + "if(!fclose($h)){fread($h,1);}}", 0},
		{"failure comparison", "function f($p){" + open + "if(fclose($h)===false){fread($h,1);}}", 0},
		{"failure else", "function f($p){" + open + "if(fclose($h)){echo 1;}else{fread($h,1);}}", 0},
		{"failure after guard", "function f($p){" + open + "if(fclose($h)){return;}fread($h,1);}", 0},
		{"failure ternary", "function f($p){" + open + "return fclose($h)?null:fread($h,1);}", 0},
		{"failure short-circuit", "function f($p){" + open + "fclose($h) or fread($h,1);}", 0},
		{"stored failure", "function f($p){" + open + "$ok=fclose($h);if($ok===false){fread($h,1);}}", 0},
		{"stored failure at file scope", "$h=fopen('a','rb');if($h!==false){$ok=fclose($h);if(!$ok){fread($h,1);}}", 0},
		{"captured by reference before close", "function f($p){" + open + "$g=function()use(&$h,$p){$h=fopen($p,'rb');};fclose($h);$g();fread($h,1);}", 0},
		{"captured by reference after close", "function f($p){" + open + "fclose($h);$g=function()use(&$h,$p){$h=fopen($p,'rb');};$g();fread($h,1);}", 0},
		{"captured by reference at file scope", "$h=fopen('a','rb');if($h!==false){$g=function()use(&$h){$h=fopen('b','rb');};fclose($h);$g();fread($h,1);}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			file := syntax.Parse("test.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(file.Errors) != 0 {
				t.Fatalf("parse: %+v", file.Errors)
			}
			if got := engine.Analyze(file); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
