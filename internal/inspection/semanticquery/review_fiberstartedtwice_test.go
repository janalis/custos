package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewFiberStartProbe dispatches on method calls only, as the rule package does.
type reviewFiberStartProbe struct{ nativeCheckProbe }

func (reviewFiberStartProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewFiberStartedTwice(t *testing.T) {
	const fiber = "$f=new Fiber(fn()=>Fiber::suspend());"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"second start", fiber + "$f->start();$f->start();", 1},
		{"second start in function", "function f(){" + fiber + "$f->start();$f->start();}", 1},
		{"third start reported too", fiber + "$f->start();$f->start();$f->start();", 2},
		{"unrelated statements between", fiber + "$f->start();$x=[1];foreach($x as $k=>$v){echo $v;}try{g();}catch(E $e){}$f->start();", 1},
		{"nested scope assignment between", fiber + "$f->start();$g=function(){$f=1;};$f->start();", 1},
		{"callable before real start", fiber + "$s=$f->start(...);$f->start();", 0},
		{"callable after real start", fiber + "$f->start();$s=$f->start(...);", 0},
		{"callable between real starts", fiber + "$f->start();$s=$f->start(...);$f->start();", 1},
		{"short destructuring rebinds", fiber + "$f->start();[$f]=[new Fiber(fn()=>2)];$f->start();", 0},
		{"list destructuring rebinds", fiber + "$f->start();list($f)=[new Fiber(fn()=>2)];$f->start();", 0},
		{"keyed destructuring rebinds", fiber + "$f->start();['a'=>$f]=['a'=>new Fiber(fn()=>2)];$f->start();", 0},
		{"foreach value rebinds", fiber + "$f->start();foreach([new Fiber(fn()=>3)] as $f){$f->start();}", 0},
		{"foreach key rebinds", fiber + "$f->start();foreach($m as $f=>$v){$f->start();}", 0},
		{"foreach rebinds before later start", fiber + "$f->start();foreach($m as $f){}$f->start();", 0},
		{"catch variable rebinds", fiber + "$f->start();try{g();}catch(E $f){}$f->start();", 0},
		{"catch without variable", fiber + "$f->start();try{g();}catch(E){}$f->start();", 1},
		{"plain reassignment", fiber + "$f->start();$f=new Fiber(fn()=>2);$f->start();", 0},
		{"unreachable second start", "function f(){" + fiber + "$f->start();return;$f->start();}", 0},
		{"incomplete syntax", fiber + "$f->start();$f->start();if(", 0},
		{"no arguments list is not a callable", fiber + "$f->start(1);$f->start(2);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewFiberStartProbe{nativeCheckProbe{"FiberStartedTwice", CheckFiberStartedTwice}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"FiberStartedTwice"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
