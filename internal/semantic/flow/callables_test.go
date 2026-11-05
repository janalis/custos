package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestKnownCallableTargets(t *testing.T) {
	for _, tc := range []struct {
		source       string
		known, taint bool
	}{
		{`function raw($x){return $x;} $f=raw(...);probe($f($_GET['x']));`, true, true},
		{`function esc($x){return htmlspecialchars($x);} $f=esc(...);probe($f($_GET['x']));`, true, false},
		{`namespace A; function raw($x){return $x;} $f=raw(...);probe($f($_GET['x']));`, true, true},
		{`function raw($x){return $x;} $f='raw';probe($f($_GET['x']));`, true, true},
		{`function raw($x){return $x;} probe(call_user_func('raw',$_GET['x']));`, true, true},
		{`function esc($x){return htmlspecialchars($x);} probe(call_user_func('esc',$_GET['x']));`, true, false},
		{`function esc($x){return htmlspecialchars($x);} $f=esc(...);probe(call_user_func($f,$_GET['x']));`, true, false},
		{`function esc($x){return htmlspecialchars($x);} probe(call_user_func(callback:'esc',x:$_GET['x']));`, true, false},
		{`function esc($x){return htmlspecialchars($x);} probe(call_user_func_array('esc',[$_GET['x']]));`, true, false},
		{`function esc($x){return htmlspecialchars($x);} $args=['x'=>$_GET['x']];probe(call_user_func_array('esc',$args));`, true, false},
		{`$f=fn($x)=>htmlspecialchars($x);probe(call_user_func($f,$_GET['x']));`, true, false},
		{`$f=fn($x)=>htmlspecialchars($x);probe(call_user_func_array($f,[$_GET['x']]));`, true, false},
		{`function raw($x){return $x;} probe(call_user_func('raw',other:$_GET['x']));`, true, true},
		{`function &raw($x){return $x;} $f=raw(...);probe($f($_GET['x']));`, false, false},
		{`function raw(&$x){return $x;} $f=raw(...);probe($f($_GET['x']));`, false, false},
		{`function raw($x){yield $x;} $f=raw(...);probe($f($_GET['x']));`, false, false},
		{`if($c){function raw($x){return $x;}}$f=raw(...);probe($f($_GET['x']));`, false, false},
		{`function raw($x){return $x;}$f=raw(...);$alias=&$f;mutate($alias);probe($f($_GET['x']));`, false, false},
		{`$f=htmlspecialchars(...);probe($f($_GET['x']));`, true, false},
		{`$f=missing(...);probe($f($_GET['x']));`, false, false},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		value := e.Value(x)
		if value.Complete != tc.known || value.Complete && e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v taint=%v", tc.source, value, e.Tainted(x, HTML))
		}
	}
}

func TestCallableCreationVersionsAndSinks(t *testing.T) {
	e, f := summarized(t, `function send($x){header($x);} $f=send(...);$f($_GET['x']);`)
	var call *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n, ok := n.(*syntax.FuncCall); ok {
			if _, dynamic := n.Name.(*syntax.Variable); dynamic {
				call = n
			}
		}
		return true
	})
	sinks := e.Sinks(call)
	if len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("sinks=%+v", sinks)
	}
	r := &scopeResult{}
	e.types.PHP = phpversion.PHP80
	e.captureCallable(call, Value{}, &frame{}, r)
	if len(r.callables) != 0 {
		t.Fatal("first-class target on old PHP")
	}
	e.types.PHP = phpversion.Max
	e.captureCallable(&syntax.StaticCall{}, Value{}, &frame{}, r)
	if len(r.callables) != 0 {
		t.Fatal("unsupported first-class target")
	}
}

func TestCallableIncompleteArguments(t *testing.T) {
	for _, source := range []string{
		`function esc($x){return htmlspecialchars($x);} probe(call_user_func(...$args));`,
		`function esc($x){return htmlspecialchars($x);} probe(call_user_func(...['esc',$_GET['x']]));`,
		`function esc($x){return htmlspecialchars($x);} probe(call_user_func_array('esc'));`,
	} {
		e, f := summarized(t, source)
		call := argument(callsNamed(f, "probe")[0]).(*syntax.FuncCall)
		scope := e.scope(syntax.EnclosingVariableScope(call))
		var values []Value
		for _, called := range scope.calls {
			if called.Node == call {
				values = called.Arguments
			}
		}
		if _, _, known := e.invokeCallable(call, values, &frame{}, scope); known {
			t.Fatal("incomplete binding established callback proof")
		}
	}
	e, f := summarized(t, `function esc($x){return htmlspecialchars($x);} probe(call_user_func(x:$_GET['x'],callback:'esc'));`)
	if e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
		t.Fatal("named callback order lost escaping")
	}
}

func BenchmarkKnownCallableFlow(b *testing.B) {
	for _, source := range []string{`function esc($x){return htmlspecialchars($x);}$f=esc(...);probe($f($_GET['x']));`, `function esc($x){return htmlspecialchars($x);}probe(call_user_func_array('esc',[$_GET['x']]));`} {
		b.Run(source, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				e, f := summarized(b, source)
				e.Value(argument(callsNamed(f, "probe")[0]))
			}
		})
	}
}
