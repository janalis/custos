package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestNativeCallableContracts(t *testing.T) {
	for _, tc := range []struct {
		source               string
		html, single, double bool
	}{
		{`$f=strtolower(...);probe($f($_GET['x']));`, true, true, true},
		{`probe(call_user_func('htmlspecialchars',$_GET['x']));`, false, false, false},
		{`$f=htmlspecialchars(...);probe($f(string:$_GET['x'],flags:ENT_COMPAT));`, false, true, false},
		{`probe(call_user_func_array('htmlspecialchars',[$_GET['x'],ENT_COMPAT]));`, false, true, false},
		{`namespace A;function htmlspecialchars($x){return $x;}$f=htmlspecialchars(...);probe($f($_GET['x']));`, true, true, true},
		{`function f($unknown){$f=htmlspecialchars(...);probe($f($_GET['x'],$unknown));}`, false, true, true},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		if !e.Value(x).Complete || e.Tainted(x, HTML) != tc.html || e.Tainted(x, HTMLSingleQuoted) != tc.single || e.Tainted(x, HTMLDoubleQuoted) != tc.double {
			t.Fatalf("%s: %+v", tc.source, e.Value(x))
		}
	}
}

func TestNativeCallableArgumentEvaluation(t *testing.T) {
	e, f := summarized(t, `$f=strtolower(...);$i=0;$f($i++);probe($i);`)
	if low, high, known := e.IntRange(argument(callsNamed(f, "probe")[0])); !known || low != 1 || high != 1 {
		t.Fatalf("argument re-evaluated: %d %d %v", low, high, known)
	}
	for _, source := range []string{
		`$close=fclose(...);$h=tmpfile();if(!$close($h)){probe($h);}`,
		`$h=tmpfile();if(!call_user_func('fclose',$h)){probe($h);}`,
	} {
		e, f = summarized(t, source)
		call := callsNamed(f, "probe")[0]
		if _, closed := e.StateBefore(call, argument(call), "fclose"); closed {
			t.Fatal("indirect failed close became proof")
		}
	}
}

func TestNativeCallableSinksAndBounds(t *testing.T) {
	e, f := summarized(t, `function wrap($x){$f=header(...);$f($x);}wrap($_GET['x']);`)
	if sinks := e.Sinks(callsNamed(f, "wrap")[0]); len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("native sinks=%+v", sinks)
	}
	e, _ = environment(t, ``, nil)
	if _, _, known := e.invokeNativeCallable(&syntax.StaticCall{}, nil, nil, "strtolower", &frame{}, &scopeResult{}); known {
		t.Fatal("invalid call target")
	}
}

func TestNativeCallableRefusesInvalidBindings(t *testing.T) {
	for _, source := range []string{
		`$f=strtolower(...);probe($f());`,
		`$f=compact(...);probe($f('x'));`,
	} {
		e, file := summarized(t, source)
		if e.Value(argument(callsNamed(file, "probe")[0])).Complete {
			t.Fatal(source)
		}
	}
	e, file := summarized(t, `$f=htmlspecialchars(...);probe($f(string:$_GET['x'],double_encode:false));`)
	if e.Tainted(argument(callsNamed(file, "probe")[0]), HTML) {
		t.Fatal("omitted optional slot lost sanitizer")
	}
}

func TestNativeCallableBudget(t *testing.T) {
	e, file := environment(t, `strtolower('x');`, nil)
	r := &scopeResult{complete: true, operations: MaxTransfers}
	if _, _, known := e.invokeNativeCallable(callsNamed(file, "strtolower")[0], nil, nil, "strtolower", &frame{}, r); known || r.complete {
		t.Fatal("native replay ignored budget")
	}
}
