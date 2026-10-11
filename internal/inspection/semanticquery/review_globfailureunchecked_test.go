package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewGlobProbe dispatches on the two node kinds the rule package registers.
type reviewGlobProbe struct{ nativeCheckProbe }

func (reviewGlobProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KFuncCall}
}

func TestReviewGlobFailureUnchecked(t *testing.T) {
	const get = "$p=glob($d);"
	const loop = "foreach($p as $x){echo $x;}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"direct loop", "foreach(glob($d) as $x){echo $x;}", 1},
		{"alias loop", get + loop, 1},
		{"direct count", "echo count(glob($d));", 1},
		{"alias array_map", get + "return array_map('basename',$p);", 1},
		{"named haystack", get + "return in_array(needle:$n,haystack:$p);", 1},
		{"first variadic array", get + "return array_merge($p,$q);", 1},
		{"later variadic array is not inspected", get + "return array_merge($q,$p);", 0},
		{"separator position", get + "return implode($p,',');", 0},
		{"unlisted consumer", get + "return json_encode($p);", 0},
		{"dynamic callee", get + "return $f($p);", 0},
		{"user function shadows consumer", "namespace N; function count($v){return 0;} " + get + "return count($p);", 0},
		{"fallback operand", "return count(glob($d)?:[]);", 0},

		{"strict return", get + "if($p===false){return;}" + loop, 0},
		{"falsy return", get + "if(!$p){return;}" + loop, 0},
		{"falsy return then other statements", get + "if(!$p){return;}$n=1;echo $n;" + loop, 0},
		{"empty return", get + "if(empty($p)){return;}" + loop, 0},
		{"loose false return", get + "if($p==false){return;}" + loop, 0},
		{"reversed loose false return", get + "if(false==$p){return;}" + loop, 0},
		{"not true return", get + "if($p!=true){return;}" + loop, 0},
		{"not array return", get + "if(!is_array($p)){return;}" + loop, 0},
		{"predicate compared with false", get + "if(is_array($p)===false){return;}" + loop, 0},
		{"predicate compared with true", get + "if(is_array($p)!==true){return;}" + loop, 0},
		{"bool return", get + "if(is_bool($p)){return;}" + loop, 0},
		{"disjunction return", get + "if($p===false||$q){return;}" + loop, 0},
		{"conjunction return proves nothing", get + "if(!$p&&$q){return;}" + loop, 1},
		{"guard does not leave", get + "if(!$p){echo 1;}" + loop, 1},
		{"guard on another variable", get + "if(!$q){return;}" + loop, 1},
		{"guard before reassignment", get + "if(!$p){return;}$p=glob($e);" + loop, 1},
		{"null comparison", get + "if($p===null){return;}" + loop, 1},
		{"comparison with variable", get + "if($p===$q){return;}" + loop, 1},
		{"unrelated predicate", get + "if(is_string($p)){return;}" + loop, 1},
		{"arithmetic negation", get + "if(-$q){return;}" + loop, 1},
		{"instanceof condition", get + "if($p instanceof Countable){return;}" + loop, 1},

		{"truthy branch", get + "if($p){" + loop + "}", 0},
		{"array branch", get + "if(is_array($p)){" + loop + "}", 0},
		{"iterable branch", get + "if(\\is_iterable($p)){" + loop + "}", 0},
		{"countable branch", get + "if(is_countable($p)){return count($p);}", 0},
		{"loose not false branch", get + "if($p!=false){" + loop + "}", 0},
		{"true comparison branch", get + "if($p==true){" + loop + "}", 0},
		{"conjunction branch", get + "if($q&&$p){" + loop + "}", 0},
		{"keyword conjunction branch", get + "if($q and !empty($p)){" + loop + "}", 0},
		{"disjunction branch proves nothing", get + "if($q||$p){" + loop + "}", 1},
		{"offset set branch", get + "if(isset($p[0])){" + loop + "}", 0},
		{"variable set branch", get + "if(isset($p)){" + loop + "}", 1},
		{"negated offset set branch", get + "if(!isset($p[0])){" + loop + "}", 1},
		{"else branch", get + "if(!$p){echo 1;}else{" + loop + "}", 0},
		{"else of truthy branch", get + "if($p){echo 1;}else{" + loop + "}", 1},
		{"elseif branch", get + "if($q){echo 1;}elseif($p){" + loop + "}", 0},
		{"elseif after falsy", get + "if(!$p){echo 1;}elseif($q){" + loop + "}", 0},
		{"unrelated elseif", get + "if($q){echo 1;}elseif($r){" + loop + "}", 1},
		{"use inside the condition", get + "if(count($p)>0){echo 1;}", 1},
		{"unrelated branch", get + "if($q){" + loop + "}", 1},

		{"short-circuit and", get + "return $p&&count($p)>1;", 0},
		{"short-circuit or", get + "return $p===false||count($p)===0;", 0},
		{"keyword or", get + "$r=(!$p or count($p)===0);", 0},
		{"left operand use", get + "return count($p)&&$p;", 1},
		{"wrong short-circuit polarity", get + "return $p||count($p);", 1},
		{"arithmetic operand", get + "return 1+count($p);", 1},
		{"ternary then", get + "return $p?count($p):0;", 0},
		{"ternary else", get + "return empty($p)?0:count($p);", 0},
		{"ternary wrong arm", get + "return $p?0:count($p);", 1},
		{"ternary else wrong arm", get + "return empty($p)?count($p):0;", 1},
		{"short ternary", get + "return $q?:count($p);", 1},

		{"or exit", get + "$p or exit(1);" + loop, 0},
		{"or die", get + "$p||die();" + loop, 0},
		{"and throw", get + "$p===false and throw new RuntimeException();" + loop, 0},
		{"and exit", get + "!$p&&exit;" + loop, 0},
		{"exit with wrong polarity", get + "$p and exit;" + loop, 1},
		{"or without leaving", get + "$p or print('none');" + loop, 1},
		{"coalesce exit", get + "$q ?? exit;" + loop, 1},
		{"plain statement", get + "trim($q);" + loop, 1},

		{"closure boundary", get + "if(!$p){return;}$f=function()use($p){" + loop + "};", 0},
		{"switch case list", get + "switch($q){case 1:if(!$p){break;}" + loop + "}", 0},
		{"incomplete if", get + "if(count($p))", 1},
		{"incomplete ternary", get + "return $q?count($p):", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewGlobProbe{nativeCheckProbe{"GlobFailureUnchecked", CheckGlobFailureUnchecked}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{"GlobFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			src := "<?php function f($d,$e,$q,$r,$n,$f){" + tc.src + "}"
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
