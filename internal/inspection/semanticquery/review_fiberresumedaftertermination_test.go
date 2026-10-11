package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewFiberResumedAfterTermination(t *testing.T) {
	const (
		done    = "$f=new Fiber(fn()=>1);"
		started = done + "$f->start();"
		other   = "new Fiber(function(){Fiber::suspend();})"
	)
	// The rule only subscribes to method calls.
	check := func(ctx *analysis.Context, n syntax.Node, message string) {
		if _, ok := n.(*syntax.MethodCall); ok {
			CheckFiberResumedAfterTermination(ctx, n, message)
		}
	}
	fn := func(body string) string { return "function f($xs){" + body + "}" }
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		// First-class callables create closures and run nothing.
		{"resume callable", started + "$r=$f->resume(...);", 0},
		{"start callable", done + "$s=$f->start(...);$f->resume();", 0},
		{"start callable then start", done + "$s=$f->start(...);$f->start();$f->resume();", 1},
		{"start callable invoked", done + "$s=$f->start(...);$s();$f->resume();", 0},
		// The variable receives another fiber.
		{"foreach value", started + "foreach([" + other + "] as $f){$f->start();$f->resume();}", 0},
		{"foreach key", started + "foreach($xs as $f=>$x){$f->resume();}", 0},
		{"foreach destructured value", started + "foreach($xs as [$f]){$f->resume();}", 0},
		{"array destructuring", started + "[$f]=[" + other + "];$f->start();$f->resume();", 0},
		{"list destructuring", started + "list($f)=$xs;$f->resume();", 0},
		{"catch variable", fn(started + "try{g();}catch(Exception $f){}$f->resume();"), 0},
		{"static variable", fn(started + "static $f;$f->resume();"), 0},
		{"unset", started + "unset($f);$f->resume();", 0},
		{"reference assignment", started + "$g=&$f;$g=" + other + ";$g->start();$f->resume();", 0},
		{"reference taken before allocation", "$g=&$f;" + started + "$f->resume();", 0},
		{"closure reference capture", started + "$c=function()use(&$f){$f=" + other + ";};$c();$f->resume();", 0},
		{"global declaration", fn("global $f;" + started + "$f->resume();"), 0},
		{"inside by-reference capturing closure", "$c=function()use(&$f){" + started + "$f->resume();};", 0},
		{"property receiver", "class C{public $f;function m(){$this->f=new Fiber(fn()=>1);$this->f->start();$this->f->resume();}}", 0},
		// Unrelated or earlier bindings keep the proof.
		{"plain", started + "$f->resume();", 1},
		{"in function", fn(started + "$f->resume();"), 1},
		{"in closure", "$c=function(){" + started + "$f->resume();};", 1},
		{"read only uses", started + "var_dump($f);$g=$f;$f->resume();", 1},
		{"foreach before allocation", "foreach($xs as $k=>$f){}" + started + "$f->resume();", 1},
		{"foreach over other variables", started + "foreach($xs as $k=>$x){}$f->resume();", 1},
		{"foreach without key", started + "foreach($xs as $x){}$f->resume();", 1},
		{"other destructuring", started + "[$g]=$xs;$f->resume();", 1},
		{"element assignment", started + "$xs[0]=$f->isTerminated();$f->resume();", 1},
		{"other reference", started + "$g=&$xs;$f->resume();", 1},
		{"catch without variable", fn(started + "try{g();}catch(Exception){}$f->resume();"), 1},
		{"other catch variable", fn(started + "try{g();}catch(Exception $e){}$f->resume();"), 1},
		{"other static", fn("static $n;" + started + "$f->resume();"), 1},
		{"other unset", started + "unset($xs);$f->resume();", 1},
		{"other global", fn("global $n;" + started + "$f->resume();"), 1},
		{"other reference capture", started + "$c=function()use(&$xs){};$f->resume();", 1},
		{"nested scope rebinding", started + "$c=function(){foreach([] as $f){}};$f->resume();", 1},
		{"rebinding after resume", started + "$f->resume();foreach($xs as $f){}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := auditNative(t, "FiberResumedAfterTermination", check, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
