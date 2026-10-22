package assertioncontainsrequiredsideeffect

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewAssertionContainsRequiredSideEffect(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"read only in another assert", "<?php function f($n){assert(($h=intdiv($n,2))>=0);assert($h<=$n);return 1;}", 0},
		{"read after another assert", "<?php function f($n){assert(($h=intdiv($n,2))>=0);assert($h<=$n);return $h;}", 1},
		{"user assert lookalike", "<?php namespace A;function assert($x){}function f($n){\\assert(($h=1)>0);assert($h);}", 1},
		{"preg_match matches", "<?php function f($s){assert(($m=[])===[]);preg_match('/a/',$s,$m);return $m;}", 0},
		{"second out position", "<?php function f(){assert(($l=0)===0);headers_sent($f,$l);return $l;}", 0},
		{"in-out sort", "<?php function f(){assert(($m=[3,1])!==[]);sort($m);return $m;}", 1},
		{"exec output appends", "<?php function f(){assert(($o=['a'])!==[]);exec('true',$o);return $o;}", 1},
		{"exec result code", "<?php function f(){assert(($c=0)===0);exec('true',$o,$c);return $c;}", 0},
		{"out parameter also read", "<?php function f($s){assert(($m='/a/')!=='');return preg_match($m,$s,$m);}", 1},
		{"out parameter in input position", "<?php function f($s){assert(($m='/a/')!=='');preg_match($m,$s);return 1;}", 1},
		{"named out argument", "<?php function f($s){assert(($m=[])===[]);preg_match('/a/',$s,matches:$m);return $m;}", 1},
		{"unpacked arguments", "<?php function f($a){assert(($m=[])===[]);preg_match(...$a);return $m;}", 1},
		{"first-class callable", "<?php function f(){assert(($m=[])===[]);$g=preg_match(...);return $m;}", 1},
		{"user function lookalike", "<?php namespace A;function preg_match($a,$b,$c){}function f($s){assert(($m=[])===[]);preg_match('/a/',$s,$m);return 1;}", 1},
		{"out after required read", "<?php function f($s){assert(($m=[])===[]);echo count($m);preg_match('/a/',$s,$m);return $m;}", 1},
		{"foreach value", "<?php function f($l){assert(($v=count($l))>=0);foreach($l as $v){}return $v;}", 0},
		{"foreach key", "<?php function f($l){assert(($k='n')!=='');foreach($l as $k=>$v){}return $k;}", 0},
		{"foreach list target", "<?php function f($l){assert(($a=0)===0);foreach($l as [$a,$b]){}return $a;}", 0},
		{"nested foreach target", "<?php function f($l,$c){assert(($v=0)===0);if($c){foreach($l as $v){}}return $v;}", 0},
		{"foreach over the local", "<?php function f(){assert(($v=[1])!==[]);foreach($v as $v){}return 1;}", 1},
		{"foreach other target", "<?php function f($l){assert(($v=1)>0);foreach($l as $w){}return $v;}", 1},
		{"read before foreach", "<?php function f($l){assert(($v=1)>0);echo $v;foreach($l as $v){}}", 1},
		{"catch variable", "<?php function f($t){assert(($e=null)===null);try{$t();}catch(Exception $e){}return $e;}", 0},
		{"other catch variable", "<?php function f($t){assert(($e=1)>0);try{$t();}catch(Exception $x){}return $e;}", 1},
		{"catch without variable", "<?php function f($t){assert(($e=1)>0);try{$t();}catch(Exception){}return $e;}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewProbe{id: "AssertionContainsRequiredSideEffect", kinds: []syntax.NodeKind{syntax.KFuncCall}, check: func(ctx *analysis.Context, n syntax.Node) {
				check(ctx, n, "finding")
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("case.php", []byte(tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v\n%s", len(got), tc.want, got, tc.src)
			}
		})
	}
}

type reviewProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p reviewProbe) ID() string { return p.id }

func (p reviewProbe) Semantic() {}

func (p reviewProbe) Flow() {}

func (p reviewProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p reviewProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
