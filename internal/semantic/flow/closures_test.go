package flow

import (
	"fmt"
	"testing"

	"custos/internal/php/syntax"
)

func TestLocalClosureInvocation(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		known, taint bool
	}{
		{"arrow", `$f=fn($x)=>$x; probe($f($_GET['x']));`, true, true},
		{"escaping", `$f=fn($x)=>htmlspecialchars($x); probe($f($_GET['x']));`, true, false},
		{"literal invocation", `probe((fn($x)=>$x)($_GET['x']));`, true, true},
		{"closure", `$f=function($x){return $x;}; probe($f($_GET['x']));`, true, true},
		{"capture snapshot", `$x=$_GET['x']; $f=fn()=>$x; $x='safe'; probe($f());`, true, true},
		{"use snapshot", `$x=$_GET['x']; $f=function()use($x){return $x;}; $x='safe'; probe($f());`, true, true},
		{"clean capture", `$x='safe'; $f=fn()=>$x; $x=$_GET['x']; probe($f());`, true, false},
		{"shadow", `$x=$_GET['x']; $f=fn($x)=>$x; probe($f('safe'));`, true, false},
		{"named", `$f=fn($x,$y)=>$y; probe($f(y:$_GET['x'],x:'safe'));`, true, true},
		{"spread", `$f=fn($x)=>$x; probe($f(...[$_GET['x']]));`, true, true},
		{"no return", `$f=function(){}; probe($f());`, true, false},
		{"reference capture", `$x=$_GET['x'];$f=function()use(&$x){return $x;};probe($f());`, false, false},
		{"reference return", `$f=function &($x){return $x;};probe($f($_GET['x']));`, false, false},
		{"reference arrow", `$f=fn&($x)=>$x;probe($f($_GET['x']));`, false, false},
		{"reference parameter", `$f=fn(&$x)=>$x;probe($f($_GET['x']));`, false, false},
		{"variadic", `$f=fn(...$x)=>$x;probe($f($_GET['x']));`, true, true},
		{"default", `$f=fn($x='safe')=>$x;probe($f());`, true, false},
		{"unknown", `probe($f($_GET['x']));`, false, false},
		{"bad binding", `$f=fn($x)=>$x;probe($f(other:$_GET['x']));`, false, false},
		{"generator", `$f=function($x){yield $x;};probe($f($_GET['x']));`, false, false},
		{"unresolved body", `$f=fn($x)=>unknown($x);probe($f($_GET['x']));`, false, false},
		{"reassigned", `$f=fn($x)=>$x; $f=unknown();probe($f($_GET['x']));`, false, false},
		{"branch ambiguity", `$f=$c?fn($x)=>$x:fn($x)=>'safe';probe($f($_GET['x']));`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := summarized(t, tc.source)
			x := argument(callsNamed(f, "probe")[0])
			v := e.Value(x)
			if v.Complete != tc.known || v.Complete && e.Tainted(x, HTML) != tc.taint {
				t.Fatalf("value=%+v taint=%v", v, e.Tainted(x, HTML))
			}
		})
	}
}

func TestClosureSinks(t *testing.T) {
	e, f := summarized(t, `function outer($x){$f=function($v){header($v);};$f($x);} outer($_GET['x']);`)
	call := callsNamed(f, "outer")[0]
	sinks := e.Sinks(call)
	if len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("sinks=%+v", sinks)
	}
	// An ordinary query of a closure body must not retain a caller's values.
	var closure *syntax.Closure
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n, ok := n.(*syntax.Closure); ok {
			closure = n
		}
		return true
	})
	if e.scope(closure).values[argument(callsNamed(f, "header")[0])].Sources[0].Kind != "parameter" {
		t.Fatal("invocation contaminated symbolic scope")
	}
}

func BenchmarkLocalClosure(b *testing.B) {
	for b.Loop() {
		e, f := summarized(b, `$x=$_GET['x'];$f=fn($y)=>$x.$y;probe($f('a'));`)
		e.Value(argument(callsNamed(f, "probe")[0]))
	}
}

func TestClosureRepeatedCaptureAndCache(t *testing.T) {
	e, f := summarized(t, `$x=$_GET['x'];while($c){$f=function()use($x){if($c){return $x;}return 'safe';};$x='safe';probe($f());}`)
	x := argument(callsNamed(f, "probe")[0])
	if !e.Value(x).Complete || !e.Tainted(x, HTML) {
		t.Fatalf("joined captures: %+v", e.Value(x))
	}
	// Query the ordinary body first, then invoke it from its creator.
	e, f = environment(t, `$f=fn($x)=>$x;probe($f($_GET['x']));`, nil)
	var arrow *syntax.ArrowFunction
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n, ok := n.(*syntax.ArrowFunction); ok {
			arrow = n
		}
		return true
	})
	before := e.scope(arrow)
	if !e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) || e.scope(arrow) != before {
		t.Fatal("scope cache was not restored")
	}
}

func TestClosureCaptureRecoveryAndBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	r := &scopeResult{}
	f := frame{vars: map[string]Value{"x": {Complete: true, replaced: true}}}
	capture := &syntax.Closure{Uses: []*syntax.ClosureUse{{Var: &syntax.Variable{Name: "x"}}}}
	e.captureClosure(capture, &f, r)
	if r.closures[capture]["x"].Complete {
		t.Fatal("stale capture")
	}
	capture = &syntax.Closure{Uses: []*syntax.ClosureUse{{}}}
	e.captureClosure(capture, &f, r)
	if _, ok := r.closures[capture]; ok {
		t.Fatal("malformed capture")
	}
	uses := make([]*syntax.ClosureUse, 65)
	for i := range uses {
		uses[i] = &syntax.ClosureUse{Var: &syntax.Variable{Name: "x"}}
	}
	capture = &syntax.Closure{Uses: uses}
	e.captureClosure(capture, &f, r)
	if _, ok := r.closures[capture]; ok {
		t.Fatal("oversized capture")
	}
}

func TestClosureInvocationBoundaries(t *testing.T) {
	for _, body := range []string{
		`$f=function(){static $x='safe';$old=$x;$x=$_GET['x'];return $old;};$f();probe($f());`,
		`$f=fn($x)=>eval($x);probe($f($_GET['x']));`,
	} {
		e, f := summarized(t, body)
		if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
			t.Fatal("stateful or incomplete closure accepted")
		}
		if len(e.Gaps()) == 0 {
			t.Fatal("missing incomplete status")
		}
	}
}

func TestClosureCaptureBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	closure := &syntax.Closure{}
	for _, r := range []*scopeResult{{operations: MaxTransfers}, {retained: MaxRetained}} {
		e.captureClosure(closure, &frame{}, r)
		if _, ok := r.closures[closure]; ok {
			t.Fatal("capture budget exceeded")
		}
	}
	code := "$f=fn("
	for i := 0; i < 65; i++ {
		if i > 0 {
			code += ","
		}
		code += fmt.Sprintf("$p%d", i)
	}
	code += ")=>'safe';probe($f());"
	e, f := summarized(t, code)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("oversized parameter list accepted")
	}
}

func TestClosureDefaultAndPackBinding(t *testing.T) {
	for _, tc := range []struct {
		source       string
		known, taint bool
	}{
		{`$f=fn($x='safe')=>$x;probe($f($_GET['x']));`, true, true},
		{`$f=fn($x=new Unknown)=>$x;probe($f($_GET['x']));`, true, true},
		{`$f=fn($x=new Unknown)=>$x;probe($f());`, false, false},
		{`$f=fn($x=['v'=>'safe'])=>$x['v'];probe($f());`, true, false},
		{`$f=fn($x=-1)=>$x;probe($f());`, true, false},
		{`$f=fn($x=null)=>$x;probe($f());`, true, false},
		{`$f=fn($x=true)=>$x;probe($f());`, true, false},
		{`$f=fn($x=Missing)=>$x;probe($f());`, false, false},
		{`$f=fn($x='safe',$y)=>$y;probe($f(y:$_GET['x']));`, false, false},
		{`$f=fn(...$xs)=>$xs[0];probe($f($_GET['x'],'safe'));`, true, true},
		{`$f=fn(...$xs)=>$xs[1];probe($f($_GET['x'],'safe'));`, true, false},
		{`$f=fn(...$xs)=>$xs['key'];probe($f(key:$_GET['x']));`, true, true},
		{`$f=fn($first,...$xs)=>$xs[0];probe($f('safe',$_GET['x']));`, true, true},
		{`$f=fn(...$xs)=>$xs[0];probe($f(...[$_GET['x']]));`, true, true},
		{`$f=fn(...$xs)=>$xs;probe($f());`, true, false},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		v := e.Value(x)
		if v.Complete != tc.known || v.Complete && e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v taint=%v", tc.source, v, e.Tainted(x, HTML))
		}
	}
}

func TestClosureDefaultBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	for _, source := range []string{`fn($x=[1.5=>'x'])=>$x`, `fn($x=[...$other])=>$x`, `fn($x=[new Unknown])=>$x`} {
		_, file := environment(t, "probe("+source+");", nil)
		arrow := argument(callsNamed(file, "probe")[0]).(*syntax.ArrowFunction)
		if _, known := e.closureDefault(arrow.Params[0].Default, &scopeResult{}); known {
			t.Fatal("effectful default accepted")
		}
	}
	if primitiveDefault(&syntax.Literal{}, 0, new(int)) {
		t.Fatal("default budget ignored")
	}
	r := &scopeResult{operations: MaxTransfers}
	if _, known := e.closureDefault(&syntax.Literal{LitKind: syntax.LitInt, Raw: "1"}, r); known {
		t.Fatal("caller budget ignored")
	}
}

func TestClosureInvocationDepth(t *testing.T) {
	expr := `$x`
	for i := 0; i < MaxDepth+1; i++ {
		expr = `(fn($x)=>` + expr + `)($x)`
	}
	e, f := summarized(t, `function f($x){probe(`+expr+`);}`)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("invocation depth bound ignored")
	}
}
