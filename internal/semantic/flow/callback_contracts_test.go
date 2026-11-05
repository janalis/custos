package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestCallbackResultContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`$xs=array_map(fn()=>$_GET['x'],['safe']);probe($xs[0]);`, true},
		{`function reject($x){return false;}$xs=array_filter([$_GET['x']],'reject');probe($xs);`, false},
		{`function factory(){return fn($x)=>false;}$cb=factory();$xs=array_filter([$_GET['x']],$cb);probe($xs);`, false},
		{`$value=array_reduce(['safe'],fn($carry,$entry)=>$_GET['x'],'');probe($value);`, true},
		{`$xs=array_filter(['00'],fn(int $value)=>$value);probe($xs);`, false},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
	e, file := summarized(t, `function wrap($x){array_reduce([$x],fn($carry,$entry)=>header($entry),'');}wrap($_GET['x']);`)
	if sinks := e.Sinks(callsNamed(file, "wrap")[0]); len(sinks) != 1 || !tainted(sinks[0].Argument) {
		t.Fatalf("reduce sinks=%+v", sinks)
	}
}

func TestFiniteReductionContracts(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`$v=array_reduce([],fn($c,$x)=>$_GET['x'],'safe');probe($v);`, false},
		{`function constantResult(){return false;}$v=array_reduce([1,2],'constantResult',false);probe($v);`, false},
		{`$v=array_reduce([],fn($c,$x)=>$_GET['x']);probe($v);`, false},
		{`$v=array_reduce(['safe',$_GET['x']],fn($c,$x)=>$x,'');probe($v);`, true},
		{`$v=array_reduce([$_GET['x'],'safe'],fn($c,$x)=>$x,'');probe($v);`, false},
		{`function last($c,$x){return $x;}$v=array_reduce(['safe',$_GET['x']],'last','');probe($v);`, true},
		{`function factory(){return fn($c,$x)=>$x;}$v=array_reduce(['safe',$_GET['x']],factory(),'');probe($v);`, true},
		{`$v=array_reduce(['safe'],fn($c,$x)=>$c,$_GET['x']);probe($v);`, true},
		{`$v=array_reduce(['safe'],fn($c,$x)=>['dirty'=>$_GET['x']]);probe($v['dirty']);`, true},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		v := e.Value(x)
		if !v.Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
	for _, source := range []string{
		`$v=array_reduce($unknown,fn($c,$x)=>$x,'safe');probe($v);`,
		`$v=array_reduce(['a'],fn(&$c,$x)=>$x,'safe');probe($v);`,
		`$v=array_reduce(['a'],unknown(...),'safe');probe($v);`,
		`$v=array_reduce(['a'],fn($c,$x,...$rest)=>$x,'safe');probe($v);`,
	} {
		e, f := summarized(t, source)
		if v := e.Value(argument(callsNamed(f, "probe")[0])); v.elements != nil {
			t.Fatalf("unsupported binding retained a shape: %s", source)
		}
	}
}

func TestSurplusCallbackGuards(t *testing.T) {
	for _, source := range []string{
		`$cb=fn()=>$_GET['x'];$v=$cb('safe');probe($v);`,
		`function user(){return $_GET['x'];}$v=array_map('user',['safe']);probe($v[0]);`,
		`function factory(string $x){return fn()=>$x;}$cb=factory($_GET['x']);$v=array_map($cb,['safe']);probe($v[0]);`,
		`final class R{static function run(){return $_GET['x'];}}$v=array_map(R::run(...),['safe']);probe($v[0]);`,
	} {
		e, f := summarized(t, source)
		x := argument(callsNamed(f, "probe")[0])
		if !e.Value(x).Complete || !e.Tainted(x, HTML) {
			t.Fatalf("surplus output lost: %s", source)
		}
	}
	for _, source := range []string{
		`$v=array_map(fn()=>func_num_args(),['safe']);probe($v);`,
		`function user(){return func_get_arg(0);}$v=array_map('user',['safe']);probe($v);`,
		`$v=array_map(fn()=>get_defined_vars(),['safe']);probe($v);`,
		`$v=array_map(fn()=>$_GET['x'],[new stdClass]);probe($v);`,
		`$cb=fn()=>$_GET['x'];$v=$cb(extra:'safe');probe($v);`,
	} {
		e, f := summarized(t, source)
		if v := e.Value(argument(callsNamed(f, "probe")[0])); v.elements != nil {
			t.Fatalf("unsupported binding retained a shape: %s", source)
		}
	}
}

func TestReductionRefusalPreservesPriorSinks(t *testing.T) {
	e, f := summarized(t, `$v=array_reduce(['a'],fn(&$carry,$entry)=>$entry,'safe');`)
	at := callsNamed(f, "array_reduce")[0]
	scope := e.scope(syntax.EnclosingVariableScope(at))
	values := []Value{scope.values.at(at.Args.Args[0].(*syntax.Arg).Value), scope.values.at(at.Args.Args[1].(*syntax.Arg).Value), scope.values.at(at.Args.Args[2].(*syntax.Arg).Value)}
	prior := Sink{Name: "prior", Argument: Value{Complete: true}}
	r := &scopeResult{closureSinks: map[syntax.Node][]Sink{at: {prior}}}
	if _, known := e.reducedCallback(at, at.Args, values, &frame{}, r); known {
		t.Fatal("reference callback accepted")
	}
	if got := r.closureSinks[at]; len(got) != 1 || got[0].Name != "prior" {
		t.Fatal("failed reduction destroyed earlier sinks")
	}
	if _, known := e.reducedCallback(&syntax.MethodCall{}, nil, nil, &frame{}, r); known {
		t.Fatal("method treated as builtin reduction")
	}
	if _, known := e.reducedCallback(at, nil, nil, &frame{}, r); known {
		t.Fatal("missing reduction arguments accepted")
	}
}
