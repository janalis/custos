package aeaddecryptiontaglengthunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewSafety pins the accepted spellings of the exact-length proof and
// the shapes that must keep reporting.
func TestReviewSafety(t *testing.T) {
	const open = `openssl_decrypt($c,"aes-256-gcm",$k,0,$iv,`
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"constant first", `$tag=$_POST["tag"];if(16!==strlen($tag)){return;}` + open + `$tag);`, 0},
		{"element measured", `if(strlen($_POST["tag"])!==16){return;}` + open + `$_POST["tag"]);`, 0},
		{"element measured, variable used", `$tag=$_POST["tag"];if(16!==strlen($_POST["tag"])){return;}` + open + `$tag);`, 0},
		{"ternary true arm", `$tag=$_POST["tag"];$p=strlen($tag)===16?` + open + `$tag):false;`, 0},
		{"ternary false arm", `$p=16!==strlen($_GET["t"])?false:` + open + `$_GET["t"]);`, 0},
		{"ternary true arm, element measured", `$tag=$_POST["tag"];$p=strlen($_POST["tag"])===16?` + open + `$tag):false;`, 0},
		{"earlier exit through endless loop", `$tag=$_POST["tag"];if(16!==strlen($tag)){while(true){sleep(1);}}$x=1;` + open + `$tag);`, 0},
		// The flow reads a decimal length only and does not leave through a
		// `declare` block: the rule ties these guards to the tag variable itself.
		{"hexadecimal length, guard before loop", `$tag=$_POST["tag"];if(0x10!==strlen($tag)){return;}$x=1;while($n--){` + open + `$tag);}`, 0},
		{"earlier exit inside a declare block", `$tag=$_POST["tag"];if(16!==strlen($tag)){declare(ticks=1){return;}}$x=1;while($n--){` + open + `$tag);}`, 0},
		{"hexadecimal length, variable rebound before the loop", `$tag=$_POST["tag"];if(0x10!==strlen($tag)){return;}$tag=$_GET["t"];while($n--){` + open + `$tag);}`, 1},
		{"declare block that falls through", `$tag=$_POST["tag"];if(16!==strlen($tag)){declare(ticks=1){echo 1;}}$x=1;` + open + `$tag);`, 1},
		{"if body", `if(16===strlen($_POST["tag"])){` + open + `$_POST["tag"]);}`, 0},
		{"else body", `if(16!==strlen($_POST["tag"])){echo 1;}else{` + open + `$_POST["tag"]);}`, 0},
		{"variable guard before loop", `$tag=$_POST["tag"];if(16!==strlen($tag)){return;}while($n--){` + open + `$tag);}`, 0},
		{"guard several statements earlier", `$tag=$_POST["tag"];if(16!==strlen($tag)){return;}$x=1;$y=2;` + open + `$tag);`, 0},
		{"element guard several statements earlier", `if(16!==strlen($_POST["tag"])){return;}$x=1;if($x){echo 1;` + open + `$_POST["tag"]);}`, 0},
		{"variable rebound after the guard", `$tag=$_POST["tag"];if(16!==strlen($tag)){return;}$tag=$_GET["t"];echo 1;` + open + `$tag);`, 1},
		{"variable rebound in the tag statement", `$tag=$_POST["tag"];if(16!==strlen($tag)){return;}$x=1;if($tag=$_GET["t"]){` + open + `$tag);}`, 1},
		{"element guard before loop", `if(16!==strlen($_POST["tag"])){return;}for(;;){` + open + `$_POST["tag"]);$_POST["tag"]=$c;}`, 1},
		{"element guard before do loop", `if(16!==strlen($_POST["tag"])){return;}do{` + open + `$_POST["tag"]);}while($n--);`, 1},
		{"other element", `if(16!==strlen($_POST["mac"])){return;}` + open + `$_POST["tag"]);`, 1},
		{"other superglobal", `if(16!==strlen($_GET["tag"])){return;}` + open + `$_POST["tag"]);`, 1},
		{"unrelated variable", `if(16!==strlen($other)){return;}` + open + `$_POST["tag"]);`, 1},
		{"other length", `if(12!==strlen($_POST["tag"])){return;}` + open + `$_POST["tag"]);`, 1},
		{"other function", `if(16!==is_string($_POST["tag"])){return;}` + open + `$_POST["tag"]);`, 1},
		{"loose comparison", `if(16!=strlen($_POST["tag"])){return;}` + open + `$_POST["tag"]);`, 1},
		{"wrong ternary arm", `$p=strlen($_POST["tag"])===16?false:` + open + `$_POST["tag"]);`, 1},
		{"short ternary", `$p=strlen($_POST["tag"])===16?:` + open + `$_POST["tag"]);`, 1},
		{"truthiness", `$p=$_POST["tag"]?` + open + `$_POST["tag"]):false;`, 1},
		{"non-terminating guard", `if(16!==strlen($_POST["tag"])){echo 1;}` + open + `$_POST["tag"]);`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
