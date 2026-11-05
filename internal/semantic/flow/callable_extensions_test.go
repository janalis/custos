package flow

import (
	"testing"
)

func TestEngineCallableExtensions(t *testing.T) {
	for _, tc := range []struct {
		source       string
		known, taint bool
	}{
		{`$f=$flag?'strtolower':'strtoupper';probe($f($_GET['x']));`, true, true},
		{`$f=$flag?'htmlspecialchars':'htmlentities';probe($f($_GET['x']));`, true, false},
		{`final class R{function __invoke($x){return $x;}}$f=new R;probe($f($_GET['x']));`, true, true},
		{`final class R{function __invoke($x){return htmlspecialchars($x);}}$f=new R;probe($f($_GET['x']));`, true, false},
		{`function factory(){return fn($x)=>$x;}$f=factory();probe($f($_GET['x']));`, true, true},
		{`function factory(){return fn($x)=>htmlspecialchars($x);}$f=factory();probe($f($_GET['x']));`, true, false},
		{`function factory(string $text){return fn()=> $text;}$f=factory($_GET['x']);probe($f());`, true, true},
		{`function factory(string $text){return fn()=>htmlspecialchars($text);}$f=factory($_GET['x']);probe($f());`, true, false},
		{`$xs=array_map('htmlspecialchars',[$_GET['x']]);probe($xs[0]);`, true, false},
		{`$xs=array_map('strtolower',[$_GET['x']]);probe($xs[0]);`, true, true},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		value := e.Value(x)
		if value.Complete != tc.known || value.Complete && e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, value)
		}
	}
}

func TestIndirectCloseSuccessGuards(t *testing.T) {
	for _, tc := range []struct {
		source string
		closed bool
	}{
		{`$h=tmpfile();$close=fclose(...);if($close($h)){probe($h);}`, true},
		{`$h=tmpfile();$close=fclose(...);if(!$close($h)){probe($h);}`, false},
		{`$h=tmpfile();if(call_user_func('fclose',$h)){probe($h);}`, true},
		{`$h=tmpfile();if(call_user_func_array('fclose',[$h])){probe($h);}`, true},
		{`$h=tmpfile();$old=$h;$args=[$h];$h=tmpfile();if(call_user_func_array('fclose',$args)){probe($old);}`, true},
		{`$h=tmpfile();$args=[$h];$h=tmpfile();if(call_user_func_array('fclose',$args)){probe($h);}`, false},
		{`$h=tmpfile();$close=fclose(...);$ok=$close($h);if($ok){probe($h);}`, true},
		{`$h=tmpfile();$close=fclose(...);$ok=$close($h);fread($h,1);if($ok){probe($h);}`, false},
	} {
		e, file := summarized(t, "function caller(){"+tc.source+"}")
		at := callsNamed(file, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); closed != tc.closed {
			t.Fatal(tc.source)
		}
	}
}

func TestInvokableObjectBoundaries(t *testing.T) {
	for _, source := range []string{
		`class R{function __invoke($x){return htmlspecialchars($x);}}function f(R $cb){probe($cb($_GET['x']));}`,
		`final class R{function __invoke($x){return htmlspecialchars($x);}}/** @param R $cb */function f($cb){probe($cb($_GET['x']));}`,
		`final class R{function __invoke($x){return htmlspecialchars($x);}}function f(?R $cb){probe($cb($_GET['x']));}`,
		`final class R{function &__invoke($x){return $x;}}$cb=new R;probe($cb($_GET['x']));`,
		`final class R{function __invoke($x){return htmlspecialchars($x);}}function f(R $cb){unknown($cb);probe($cb($_GET['x']));}`,
		`$cb=$flag?'htmlspecialchars':'unknown';probe($cb($_GET['x']));`,
	} {
		e, file := summarized(t, source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if value.Complete && value.Safe&HTML != 0 {
			t.Fatal(source)
		}
	}
}
