package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestFiniteFilterPreservesOriginalValues(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`$xs=array_filter(['safe',$_GET['x']],fn($v)=>true);probe($xs[0]);`, false},
		{`$xs=array_filter(['safe',$_GET['x']],fn($v)=>true);probe($xs[1]);`, true},
		{`$xs=array_filter([$_GET['x']],fn($v)=>htmlspecialchars($v));probe($xs[0]);`, true},
		{`$xs=array_filter([$_GET['x']],fn($v):string=>-0.0);probe($xs[0]);`, true},
		{`$xs=array_filter([$_GET['x']],fn($v):bool=>false);probe($xs);`, false},
		{`$xs=array_filter([$_GET['x']],function($v):string{return -0.0;});probe($xs[0]);`, true},
		{`$xs=array_filter([$_GET['x']],fn($v)=>unknown($v));probe($xs[0]);`, true},
		{`$xs=array_filter(['safe',$_GET['x']],fn($v)=>false);probe($xs);`, false},
		{`$xs=array_filter(['keep'=>'text','drop'=>'','zero'=>'0']);probe($xs['keep']);`, false},
		{`$xs=array_filter(array:['keep'=>'text','drop'=>''],callback:null);probe($xs['keep']);`, false},
		{`$xs=array_filter(['key'=>$_GET['x']],fn($k)=>$k,ARRAY_FILTER_USE_KEY);probe($xs['key']);`, true},
		{`$xs=array_filter(['key'=>$_GET['x']],fn($v,$k)=>$k,ARRAY_FILTER_USE_BOTH);probe($xs['key']);`, true},
		{`$xs=array_filter(...[['safe',$_GET['x']],'strtolower']);probe($xs[1]);`, true},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
}

func TestFiniteFilterModesAndUncertainty(t *testing.T) {
	for _, tc := range []struct {
		source         string
		known, tainted bool
	}{
		{`function wrap($x){array_filter(['key'=>$x],fn($v)=>header($v));}wrap($_GET['x']);`, true, true},
		{`function make(){return fn($v)=>header($v);}function wrap($x){$cb=make();array_filter([$x],$cb);}wrap($_GET['x']);`, true, true},
		{`function wrap($x){$cb=unknown();array_filter([$x],$cb);}wrap($_GET['x']);`, false, false},
		{`function wrap($x){$a=['safe'];$a[0]=$x;array_filter($a,'header');}wrap($_GET['x']);`, false, false},
		{`function wrap($x){array_filter(['key'=>$x],fn($v,$k)=>header($v),1);}wrap($_GET['x']);`, true, true},
		{`function wrap($x){array_filter(['key'=>$x],fn($k)=>header($k),2);}wrap($_GET['x']);`, true, false},
		{`function wrap($x,$flag){$cb=$flag?'header':'unknown';array_filter([$x],$cb);}wrap($_GET['x'],true);`, false, false},
		{`function wrap($x,$mode){array_filter([$x],'header',$mode);}wrap($_GET['x'],1);`, false, false},
		{`function wrap($x){array_filter([$x],'header',3);}wrap($_GET['x']);`, false, false},
		{`function wrap($x){array_filter(unknown(),fn($v)=>header($v));}wrap($_GET['x']);`, false, false},
	} {
		e, file := summarized(t, tc.source)
		sinks := e.Sinks(callsNamed(file, "wrap")[0])
		if (len(sinks) > 0) != tc.known || len(sinks) > 0 && tainted(sinks[0].Argument) != tc.tainted {
			t.Fatalf("%s: %+v", tc.source, sinks)
		}
	}
	e, file := environment(t, `array_filter(['safe'],'strtolower',2);array_filter(['safe'],null,0);`, nil)
	e.types.PHP = phpversion.PHP55
	r := e.scope(nil)
	for _, at := range callsNamed(file, "array_filter") {
		if _, known := e.filteredCallback(at, at.Args, r.calls[r.callIndex[at]].Arguments, &frame{}, r); known {
			t.Fatal("pre-5.6 mode established filter proof")
		}
	}
}

func TestFiniteFilterRollbackAndMutation(t *testing.T) {
	e, file := summarized(t, `array_filter(['safe'],'unknown');`)
	at := callsNamed(file, "array_filter")[0]
	r := e.scope(nil)
	r.closureSinks = map[syntax.Node][]Sink{at: {{Context: HTML}}}
	if _, known := e.filteredCallback(at, at.Args, r.calls[r.callIndex[at]].Arguments, &frame{}, r); known || len(r.closureSinks[at]) != 1 {
		t.Fatal("failed filter replaced prior sinks")
	}
	for _, source := range []string{
		`$a=['safe'];$a[$_GET['key']]=$_GET['x'];$xs=array_filter($a,fn($v)=>true);probe($xs);`,
		`$xs=array_filter([$_GET['x']],function(&$v){$v='safe';return true;});probe($xs[0]);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if v.Complete && !e.Tainted(x, HTML) {
			t.Fatalf("unsupported filter invented clean result: %s", source)
		}
	}
}

func TestPredicateParameterCoercionDoesNotInventEmptyOutput(t *testing.T) {
	e, file := summarized(t, `$xs=array_filter([-0.0],fn(string $value)=>$value);probe($xs);`)
	value := e.Value(argument(callsNamed(file, "probe")[0]))
	if value.elements != nil && value.elements.whole && !value.elements.optional && len(value.elements.items) == 0 {
		t.Fatal("unconverted falsey zero erased a truthy string predicate result")
	}
}

func TestPossibleFilterEntriesStillMapCallbackSources(t *testing.T) {
	for _, source := range []string{
		`$values=array_filter([-0.0],fn(string $value)=>$value);$mapped=array_map(fn($ignored)=>$_GET['x'],$values);probe($mapped[0]);`,
		`$values=array_filter(['safe'],fn($value)=>unknown($value));$mapped=array_map(fn($ignored)=>$_GET['x'],$values);probe($mapped[0]);`,
	} {
		e, file := summarized(t, source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !e.Tainted(x, HTML) {
			t.Fatalf("possible filter entry lost mapped callback taint: %s %+v", source, v)
		}
	}
}

func TestOptionalFilterEntriesDoNotProveMultiArrayPositions(t *testing.T) {
	e, file := summarized(t, `$values=array_filter(['safe',$_GET['x']],fn($value)=>unknown($value));$mapped=array_map(fn($a,$b)=>$a,$values,['first','second']);probe($mapped[0]);`)
	x := argument(callsNamed(file, "probe")[0])
	value := e.Value(x)
	if value.Complete && !e.Tainted(x, HTML) {
		t.Fatal("optional first entry fixed the positional mapping incorrectly")
	}
}
