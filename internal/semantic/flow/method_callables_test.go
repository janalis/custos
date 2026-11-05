package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestBoundMethodCallableTargets(t *testing.T) {
	for _, source := range []string{
		`final class R{static function raw($x){return $x;}} function f(){$cb=R::raw(...);probe($cb($_GET['x']));}`,
		`final class R{function raw($x){return $x;}} function f(R $r){$cb=$r->raw(...);probe($cb($_GET['x']));}`,
	} {
		e, f := summarized(t, source)
		x := argument(callsNamed(f, "probe")[0])
		r := e.scope(syntax.EnclosingVariableScope(x))
		if !e.Value(x).Complete || !e.Tainted(x, HTML) {
			t.Fatalf("%s: %+v methods=%+v", source, e.Value(x), r.boundMethods)
		}
	}
}

func TestBoundMethodColdPublication(t *testing.T) {
	e, f := environment(t, `final class Renderer{public static function raw($x){return $x;} public function safe($x){return htmlspecialchars($x);}} function caller(){$cb=Renderer::raw(...);probe($cb($_GET['x']));}`, nil)
	env, own := Summarize(f, e.types, nil, Options{})
	env = env.Rebase(NewSnapshot(own))
	x := argument(callsNamed(f, "probe")[0])
	scope := env.scope(syntax.EnclosingVariableScope(x))
	if !env.Value(x).Complete || !env.Tainted(x, HTML) {
		t.Fatalf("value=%+v methods=%+v summaries=%+v", env.Value(x), scope.boundMethods, own.Summaries)
	}
}

func TestMethodCallbackForms(t *testing.T) {
	for _, tc := range []struct {
		source       string
		known, taint bool
	}{
		{`final class R{function esc($x){return htmlspecialchars($x);}}function f(R $r){probe(call_user_func([$r,'esc'],$_GET['x']));}`, true, false},
		{`final class R{static function raw($x){return $x;}}probe(call_user_func(['R','raw'],$_GET['x']));`, true, true},
		{`final class R{static function raw($x){return $x;}}probe(call_user_func([R::class,'raw'],$_GET['x']));`, true, true},
		{`final class R{function raw($x){return $x;}}function f(R $r){probe(call_user_func_array([$r,'raw'],[$_GET['x']]));}`, true, true},
		{`class R{final function raw($x){return $x;}}function f(R $r){$cb=$r->raw(...);probe($cb($_GET['x']));}`, true, true},
		{`class R{private function raw($x){return $x;}function f(){$cb=$this->raw(...);probe($cb($_GET['x']));}}`, true, true},
		{`class R{function raw($x){return $x;}}function f(R $r){$cb=$r->raw(...);probe($cb($_GET['x']));}`, false, false},
		{`final class R{function raw($x){return $x;}}/** @param R $r */function f($r){$cb=$r->raw(...);probe($cb($_GET['x']));}`, false, false},
		{`final class R{function raw($x){return $x;}}function f(?R $r){$cb=$r->raw(...);probe($cb($_GET['x']));}`, false, false},
		{`final class R{function raw($x){return $x;}}function f(R $r){mutate($r);$cb=$r->raw(...);probe($cb($_GET['x']));}`, false, false},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		value := e.Value(x)
		if value.Complete != tc.known || value.Complete && e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, value)
		}
	}
}

func TestRejectedMethodCallbacksDoNotSanitize(t *testing.T) {
	for _, source := range []string{
		`final class R{static function raw($x){return htmlspecialchars($x);}}function f(){$cls='R';$cb=$cls::raw(...);probe($cb($_GET['x']));}`,
		`final class R{static function raw($x){return htmlspecialchars($x);}static function f(){$cb=static::raw(...);probe($cb($_GET['x']));}}`,
		`final class R{static function &raw($x){return $x;}}$cb=R::raw(...);probe($cb($_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}probe(call_user_func(['R'],$_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}probe(call_user_func(['object'=>'R','method'=>'raw'],$_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}probe(call_user_func(['R',unknown()],$_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}probe(call_user_func([1,'raw'],$_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}probe(call_user_func(['R','missing'],$_GET['x']));`,
		`final class R{static function raw($x){return htmlspecialchars($x);}}$cb=R::raw(...);probe($cb(wrong:$_GET['x']));`,
		`final class R{static function raw($x){return unknown($x);}}$cb=R::raw(...);probe($cb($_GET['x']));`,
	} {
		e, file := summarized(t, source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if value.Complete && value.Safe&HTML != 0 {
			t.Fatal(source)
		}
	}
}

func TestBoundMethodSinkPropagation(t *testing.T) {
	e, file := summarized(t, `final class R{static function send($x){header($x);return $x;}}function wrap($x){$f=R::send(...);return $f($x);}wrap($_GET['x']);`)
	if sinks := e.Sinks(callsNamed(file, "wrap")[0]); len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("sinks=%+v", sinks)
	}
}

func TestMethodTupleRequiresLiteralEntries(t *testing.T) {
	for _, source := range []string{
		`$r='R';probe([&$r,'raw']);`,
		`probe([...['R','raw']]);`,
		`$key=0;probe([$key=>'R',1=>'raw']);`,
		`probe([unknown(),'raw']);`,
		`function f($object){probe([$object,'raw']);}`,
	} {
		e, file := summarized(t, source)
		at := callsNamed(file, "probe")[0]
		array := argument(at).(*syntax.Array)
		r := e.scope(syntax.EnclosingVariableScope(at))
		if _, known := e.arrayMethodCallable(at, array, &frame{}, r); known {
			t.Fatal(source)
		}
	}
	e, file := environment(t, `probe(1);`, nil)
	if _, known := e.captureMethod(callsNamed(file, "probe")[0], Value{}, &frame{}); known {
		t.Fatal("non-method captured")
	}
}
