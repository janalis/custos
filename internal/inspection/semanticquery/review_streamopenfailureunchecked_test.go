package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewStreamOpenFailureUnchecked(t *testing.T) {
	const open = "$h=fopen($p,'rb');"
	const read = "fread($h,4);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"unchecked", open + read, 1},
		{"direct call argument", "fread(fopen($p,'rb'),4);", 1},
		{"property handle", "$this->h=fopen($p,'rb');fread($this->h,4);", 0},

		{"strict return", open + "if($h===false){return;}" + read, 0},
		{"falsy throw", open + "if(!$h){throw new RuntimeException();}" + read, 0},
		{"falsy return then other statements", open + "if(!$h){return;}$n=1;echo $n;" + read, 0},
		{"empty return", open + "if(empty($h)){return;}" + read, 0},
		{"loose false return", open + "if($h==false){return;}" + read, 0},
		{"reversed strict false return", open + "if(false===$h){return;}" + read, 0},
		{"not true return", open + "if($h!=true){return;}" + read, 0},
		{"not resource return", open + "if(!is_resource($h)){return;}" + read, 0},
		{"predicate compared with false", open + "if(is_resource($h)===false){return;}" + read, 0},
		{"predicate compared with true", open + "if(\\is_resource($h)!==true){return;}" + read, 0},
		{"bool return", open + "if(is_bool($h)){return;}" + read, 0},
		{"disjunction return", open + "if(!isset($h)||!$h){return;}" + read, 0},
		{"conjunction return proves nothing", open + "if(!$h&&$q){return;}" + read, 1},
		{"guard does not leave", open + "if(!$h){echo 1;}" + read, 1},
		{"guard calls unknown function", open + "if(!$h){fail();}" + read, 1},
		{"guard on another variable", open + "if(!$q){return;}" + read, 1},
		{"guard before reassignment", open + "if(!$h){return;}$h=fopen($q,'rb');" + read, 1},
		{"null comparison", open + "if($h===null){return;}" + read, 1},
		{"comparison with variable", open + "if($h===$q){return;}" + read, 1},
		{"unrelated predicate", open + "if(is_string($h)){return;}" + read, 1},
		{"arithmetic negation", open + "if(-$q){return;}" + read, 1},
		{"instanceof condition", open + "if($h instanceof Countable){return;}" + read, 1},
		{"assertion is not a guard", open + "assert($h!==false);" + read, 1},
		{"stored comparison is not a guard", open + "$ok=$h!==false;if($ok){" + read + "}", 1},
		{"loop continue", "foreach($q as $p){" + open + "if(!$h){continue;}" + read + "}", 0},
		{"loop break", "do{" + open + "if(!$h){break;}" + read + "}while(0);", 0},
		{"inside try", "try{" + open + "if(!$h){throw new Exception();}" + read + "}finally{}", 0},

		{"truthy branch", open + "if($h){" + read + "}", 0},
		{"resource branch", open + "if(is_resource($h)){" + read + "}", 0},
		{"loose not false branch", open + "if($h!=false){" + read + "}", 0},
		{"true comparison branch", open + "if($h==true){" + read + "}", 0},
		{"conjunction branch", open + "if($q&&$h){" + read + "}", 0},
		{"keyword conjunction branch", open + "if($q and !empty($h)){" + read + "}", 0},
		{"disjunction branch proves nothing", open + "if($q||$h){" + read + "}", 1},
		{"else branch", open + "if(!$h){echo 1;}else{" + read + "}", 0},
		{"else of truthy branch", open + "if($h){echo 1;}else{" + read + "}", 1},
		{"elseif branch", open + "if($q){echo 1;}elseif($h){" + read + "}", 0},
		{"elseif after falsy", open + "if(!$h){echo 1;}elseif($q){" + read + "}", 0},
		{"unrelated elseif", open + "if($q){echo 1;}elseif($r){" + read + "}", 1},
		{"use inside the condition", open + "if(fread($h,4)){echo 1;}", 1},
		{"unrelated branch", open + "if($q){" + read + "}", 1},

		{"assignment condition", "if($h=fopen($p,'rb')){" + read + "}", 0},
		{"compared assignment condition", "if(($h=fopen($p,'rb'))!==false){" + read + "}", 0},
		{"assignment condition else", "if($h=fopen($p,'rb')){echo 1;}else{" + read + "}", 1},
		{"assignment of another handle", open + "if($g=fopen($q,'rb')){" + read + "}", 1},
		{"assignment to another variable", open + "if($g=$h){" + read + "}", 1},
		{"reference assignment condition", open + "if($g=&$h){echo 1;}", 0},
		{"assignment or die", "$h=fopen($p,'rb') or die('x');" + read, 0},
		{"assignment or print", "$h=fopen($p,'rb') or print('x');" + read, 1},

		{"loop condition", open + "while($h&&!feof($h)){fgets($h);}", 0},
		{"unrelated loop condition", open + "while($q){fgets($h);}", 1},
		{"use in loop condition", open + "while(fgets($h)){echo 1;}", 1},
		{"short-circuit and", open + "return $h&&fwrite($h,'x');", 0},
		{"short-circuit or", open + "return $h===false||fread($h,1)==='';", 0},
		{"left operand use", open + "return fread($h,1)&&$h;", 1},
		{"wrong short-circuit polarity", open + "return $h||fread($h,1);", 1},
		{"concatenated operand", open + "return 'x'.fread($h,1);", 1},

		// Tests the flow facts do not model reach the syntactic guard walk.
		{"bool compared with true return", open + "if(is_bool($h)===true){return;}" + read, 0},
		{"compared predicate operand", open + "return is_resource($h)===true&&fwrite($h,'x');", 0},
		{"true comparison ternary", open + "return $h==true?fread($h,1):'';", 0},
		{"compared predicate ternary else", open + "return is_resource($h)===false?'':fread($h,1);", 0},
		{"true comparison loop condition", open + "while($h==true&&!feof($h)){fgets($h);}", 0},
		{"true comparison elseif", open + "if($q){echo 1;}elseif($h==true){" + read + "}", 0},
		{"ternary then", open + "return $h?fread($h,1):null;", 0},
		{"ternary else", open + "return $h===false?null:fread($h,1);", 0},
		{"ternary wrong arm", open + "return $h?null:fread($h,1);", 1},
		{"ternary else wrong arm", open + "return empty($h)?fread($h,1):null;", 1},
		{"short ternary", open + "return $q?:fread($h,1);", 1},

		{"or exit", open + "$h or exit(1);" + read, 0},
		{"or die", open + "$h||die();" + read, 0},
		{"or throw", open + "$h!==false||throw new Exception('x');" + read, 0},
		{"and throw", open + "$h===false and throw new RuntimeException();" + read, 0},
		{"exit with wrong polarity", open + "$h and exit;" + read, 1},
		{"or without leaving", open + "$h or print('none');" + read, 1},
		{"coalesce exit", open + "$q ?? exit;" + read, 1},
		{"short ternary throw", open + "$h?:throw new Exception('x');" + read, 0},
		{"ternary else exit", open + "$h?null:exit(1);" + read, 0},
		{"ternary then throw", open + "!$h?throw new Exception('x'):null;" + read, 0},
		{"ternary then throw with wrong polarity", open + "$h?throw new Exception('x'):null;" + read, 1},
		{"ternary without leaving", open + "$h?print('a'):print('b');" + read, 1},
		{"plain statement", open + "trim($q);" + read, 1},

		{"closure boundary", open + "if(!$h){return;}$f=function()use($h){" + read + "};", 0},
		{"switch case list", open + "switch($q){case 1:if(!$h){break;}" + read + "}", 0},
		{"incomplete if", open + "if(fread($h,4))", 1},
		{"incomplete ternary", open + "return $q?fread($h,4):", 1},
		{"incomplete loop", open + "while(fread($h,4))", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := nativeCheckProbe{"StreamOpenFailureUnchecked", CheckStreamOpenFailureUnchecked}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{"StreamOpenFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			src := "<?php function f($p,$q,$r){" + tc.src + "}"
			got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
