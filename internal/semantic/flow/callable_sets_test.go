package flow

import (
	"fmt"
	"testing"

	"custos/internal/php/syntax"
)

func TestFiniteCallableAlternatives(t *testing.T) {
	for _, tc := range []struct {
		source       string
		known, taint bool
	}{
		{`$f=$flag?fn($x)=>$x:fn($x)=>$x;probe($f($_GET['x']));`, true, true},
		{`$f=$flag?fn($x)=>htmlspecialchars($x):fn($x)=>htmlspecialchars($x);probe($f($_GET['x']));`, true, false},
		{`if($flag){$f=fn($x)=>$x;}else{$f=fn($x)=>'safe';}probe($f($_GET['x']));`, true, true},
		{`$f=$flag?strtolower(...):htmlspecialchars(...);probe($f($_GET['x']));`, true, true},
		{`function a($x){return $x;}function b($x){return $x;}$f=$flag?a(...):b(...);probe($f($_GET['x']));`, true, true},
		{`$f=$flag?fn($x)=>$x:unknown();probe($f($_GET['x']));`, false, false},
		{`$f=$flag?fn($x)=>$x:fn($x)=>unknown($x);probe($f($_GET['x']));`, false, false},
		{`$f=$flag?fn($x)=>$x:fn($y)=>$y;probe($f(x:$_GET['x']));`, false, false},
		{`$f=$flag?fn($x)=>$x:fn($x)=>$x;$alias=&$f;unknown($alias);probe($f($_GET['x']));`, false, false},
	} {
		e, f := summarized(t, tc.source)
		x := argument(callsNamed(f, "probe")[0])
		v := e.Value(x)
		if v.Complete != tc.known || v.Complete && e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
	source := `fn($x)=>$x`
	for i := 0; i < maxCallableTargets; i++ {
		source = fmt.Sprintf(`$flags[%d] ? fn($x)=>$x : (%s)`, i, source)
	}
	e, f := summarized(t, `$f=`+source+`;probe($f($_GET['x']));`)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("target fanout exceeded")
	}
}

func TestCallableSetBudgetAndSinkRollback(t *testing.T) {
	e, file := summarized(t, `$f=header(...);probe($f('text'));`)
	at := argument(callsNamed(file, "probe")[0]).(*syntax.FuncCall)
	for _, prior := range []bool{false, true} {
		r := &scopeResult{complete: true, operations: MaxTransfers, closureSinks: map[syntax.Node][]Sink{}}
		if prior {
			r.closureSinks[at] = []Sink{{Context: HTML}}
		}
		_, _, known := e.invokeCallableSet(at, Value{Expr: at.Name}, at.Args, nil, &frame{}, r)
		if known || r.complete || len(r.closureSinks[at]) != map[bool]int{false: 0, true: 1}[prior] {
			t.Fatal("budget published partial callable proof")
		}
	}
	r := &scopeResult{complete: true, closureSinks: map[syntax.Node][]Sink{at: {{Context: HTML}}}}
	_, _, known := e.invokeCallableSet(at, Value{Expr: nil}, at.Args, nil, &frame{}, r)
	if known || len(r.closureSinks[at]) != 1 {
		t.Fatal("failed target lost prior sinks")
	}
}
