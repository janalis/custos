package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestMappedCallableContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`$xs=array_map(array:[$_GET['x']],callback:'htmlspecialchars');probe($xs[0]);`, false},
		{`$xs=array_map(fn($x)=>htmlspecialchars($x),['label'=>$_GET['x']]);probe($xs['label']);`, false},
		{`function esc($x){return htmlspecialchars($x);}$xs=array_map('esc',[$_GET['x']]);probe($xs[0]);`, false},
		{`$xs=array_map('unknown',[$_GET['x']]);probe($xs[0]);`, true},
		{`$xs=array_map('htmlspecialchars',...[[$_GET['x']]]);probe($xs[0]);`, false},
		{`$xs=array_map(wrong:[$_GET['x']],callback:'htmlspecialchars');probe($xs[0]);`, true},
		{`function f($key){$a=['clean'];$a[$key]=$_GET['x'];$xs=array_map('strtolower',$a);probe($xs);}`, true},
		{`function f($flag){$a=$flag?['shared'=>'clean','left'=>$_GET['x']]:['shared'=>'clean','right'=>$_GET['x']];$xs=array_map('strtolower',$a);probe($xs);}`, true},
		{`$xs=array_map(null,[$_GET['x']]);probe($xs[0]);`, true},
		{`$xs=array_map('htmlspecialchars',[$_GET['x']],[ENT_QUOTES]);probe($xs[0]);`, false},
		{`$xs=array_map(...['htmlspecialchars',[$_GET['x']]]);probe($xs[0]);`, false},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		if !e.Value(x).Complete && tc.taint {
			continue
		}
		if !e.Value(x).Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, e.Value(x))
		}
	}
}

func TestMappedAndReturnedSinks(t *testing.T) {
	for _, source := range []string{
		`function wrap($x){array_map('header',[$x]);}wrap($_GET['x']);`,
		`function factory(){return fn($x)=>header($x);}function wrap($x){$f=factory();$f($x);}wrap($_GET['x']);`,
		`final class R{function __invoke($x){header($x);return $x;}}function wrap($x){$r=new R;$r($x);}wrap($_GET['x']);`,
	} {
		e, file := summarized(t, source)
		if sinks := e.Sinks(callsNamed(file, "wrap")[0]); len(sinks) != 1 || !tainted(sinks[0].Argument) {
			t.Fatalf("%s sinks=%+v", source, sinks)
		}
	}
}

func TestUnknownMapTargetRollsBackSinks(t *testing.T) {
	e, file := summarized(t, `function wrap($x,$flag){$cb=$flag?'header':'unknown';array_map($cb,[$x]);}wrap($_GET['x'],true);`)
	if sinks := e.Sinks(callsNamed(file, "wrap")[0]); len(sinks) != 0 {
		t.Fatalf("partial sinks=%+v", sinks)
	}
}

func TestMappedFailurePreservesPriorSinks(t *testing.T) {
	e, file := summarized(t, `array_map('unknown',[$_GET['x']]);`)
	at := callsNamed(file, "array_map")[0]
	r := e.scope(nil)
	r.closureSinks = map[syntax.Node][]Sink{at: {{Context: HTML}}}
	values := r.calls[r.callIndex[at]].Arguments
	if _, known := e.mappedCallback(at, at.Args, values, &frame{}, r); known || len(r.closureSinks[at]) != 1 {
		t.Fatal("failed mapping overwrote earlier sinks")
	}
}
