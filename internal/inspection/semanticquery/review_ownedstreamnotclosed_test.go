package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewOwnedStreamNotClosed(t *testing.T) {
	const open = "$h=fopen($p,'rb');if($h===false){return false;}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"borrowed by a reader", open + "return fgets($h);", 1},
		{"borrowed by several readers", open + "return fgets($h).ftell($h);", 1},
		{"parenthesised borrow", open + "return fgets(($h));", 1},
		{"named stream argument", open + "return fread(length:8,stream:$h);", 1},
		{"tested before the return", open + "if($h){echo 1;}if(!$h||$h==null){echo 2;}return fgets($h);", 1},
		{"tested by elseif and loops", open + "if($q){echo 1;}elseif($h){echo 2;}while($h){break;}do{echo 3;}while($h);return fgets($h);", 1},
		{"tested by a ternary condition", open + "return $h?fgets($h):'';", 1},
		{"ternary condition in parentheses", open + "return ($h)?fgets($h):'';", 1},
		{"closed on another path only", open + "if($q){fclose($h);return 1;}return fgets($h);", 1},
		{"unrelated closure and arrow", open + "$f=function()use($q){return $q;};$g=fn($x)=>$x+$q;return fgets($h);", 1},
		{"nested function with its own local", open + "function inner($h){return [$h];}return fgets($h);", 1},
		{"inside a closure scope", "return function($p){" + open + "return fgets($h);};", 1},

		{"closed by the return expression", open + "return fclose($h);", 0},
		{"read then closed in the return", open + "return fgets($h).fclose($h);", 0},
		{"wrapped in an object", open + "return new Stream($h);", 0},
		{"returned in a list", open + "return [$h,filesize($p)];", 0},
		{"returned under a key", open + "return ['handle'=>$h];", 0},
		{"handed to a method", open + "return $this->wrap($h);", 0},
		{"handed to an unknown function", open + "return wrap($h);", 0},
		{"handed to a dynamic callee", open + "return $q($h);", 0},
		{"short ternary result", open + "return $h?:null;", 0},
		{"ternary result", open + "return is_resource($h)?$h:null;", 0},
		{"captured by a returned closure", open + "return function()use($h){return fgets($h);};", 0},
		{"captured by a returned arrow function", open + "return fn()=>fgets($h);", 0},
		{"stored on the object first", open + "$this->h=$h;return fgets($h);", 0},
		{"stored in a static pool first", open + "static::$pool[]=$h;return fgets($h);", 0},
		{"stored in globals first", open + "$GLOBALS['h']=$h;return fgets($h);", 0},
		{"aliased first", open + "$copy=$h;return fgets($h);", 0},
		{"bound by reference", open + "$h=&$q;return fgets($h);", 0},
		{"registered for shutdown", open + "register_shutdown_function('fclose',$h);return fgets($h);", 0},
		{"unpacked into a builtin", open + "return fgets(...$h);", 0},
		{"cast operand", open + "return (int)$h+strlen(fgets($h));", 0},
		{"arithmetic operand", open + "return $h+strlen(fgets($h));", 0},
		{"user function shadows the reader", "namespace N; function fgets($s){return $s;} function f($p){" + open + "return fgets($h);}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := nativeCheckProbe{"OwnedStreamNotClosed", CheckOwnedStreamNotClosed}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{"OwnedStreamNotClosed"}})
			if err != nil {
				t.Fatal(err)
			}
			src := "<?php class K{function f($p,$q){" + tc.src + "}}"
			if len(tc.src) > 9 && tc.src[:9] == "namespace" {
				src = "<?php " + tc.src
			}
			got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
