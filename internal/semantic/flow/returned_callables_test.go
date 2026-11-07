package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReturnedCallableBindings(t *testing.T) {
	for _, tc := range []struct {
		source string
		taint  bool
	}{
		{`function factory(){return function($x){if($x){return $x;}return 'clean';};}$f=factory();probe($f($_GET['x']));`, true},
		{`function factory(string $x){return fn()=>$x;}$a=factory('clean');$b=factory('clean');$f=$flag?$a:$b;probe($f());`, false},
		{`function factory(){return function($x){return $x;};}$f=factory();probe($f($_GET['x']));`, true},
		{`function factory(){return fn($x='clean')=>$x;}$f=factory();probe($f());`, false},
		{`function factory(){return fn($x='clean')=>$x;}$f=factory();probe($f(x:$_GET['x']));`, true},
		{`function factory(string $x){$cb=fn()=>$x;$x='safe';return $cb;}$f=factory($_GET['x']);probe($f());`, true},
		{`function factory(string $x){return function()use($x){return $x;};}$f=factory($_GET['x']);probe($f());`, true},
		{`function factory(string $x){return fn()=>$x;}$a=factory($_GET['x']);$b=factory('clean');probe($b());`, false},
		{`function factory(string $x){return fn()=>$x;}$a=factory($_GET['x']);$b=factory('clean');$f=$flag?$a:$b;probe($f());`, true},
		{`function factory(){return fn($x)=>htmlspecialchars($x);}$f=factory();$xs=array_map($f,[$_GET['x']]);probe($xs[0]);`, false},
	} {
		e, file := summarized(t, tc.source)
		x := argument(callsNamed(file, "probe")[0])
		if !e.Value(x).Complete || e.Tainted(x, HTML) != tc.taint {
			t.Fatalf("%s: %+v", tc.source, e.Value(x))
		}
	}
}

func TestReturnedCallableProofBoundaries(t *testing.T) {
	for _, body := range []string{
		`return function()use(&$x){return $x;};`,
		`return fn(&$a)=>$a;`,
		`return fn(...$xs)=>$xs[0];`,
		`return fn($a=unknown())=>$a;`,
		`return function(){static $x;return $x;};`,
		`return function(){global $x;return $x;};`,
		`return function(){yield 'x';};`,
		`return fn()=>unknown();`,
		`return fn()=>fn()=>1;`,
		`return fn()=> $x;`,
	} {
		e, _ := summarized(t, `function factory($x){`+body+`}`)
		summary, _ := e.contract("factory")
		if summary.returned != nil {
			t.Fatal(body)
		}
	}
	e, _ := summarized(t, `function factory(){return fn(`+strings.Repeat(`$x,`, 64)+`$last)=>1;}`)
	summary, _ := e.contract("factory")
	if summary.returned != nil {
		t.Fatal("parameter bound ignored")
	}
}

func TestReturnedCallableCrossFile(t *testing.T) {
	defs := `<?php function factory(){return fn($x)=>$x;}`
	caller := `<?php function caller(){$f=factory();probe($f($_GET['x']));}`
	ix := stubsIndexForResource(t, defs, caller)
	file := syntax.Parse("definitions.php", []byte(defs), syntax.Options{Version: phpversion.Max})
	snapshot := NewSnapshot(Extract(file, ix, phpversion.Max, nil, Options{}))
	use := syntax.Parse("caller.php", []byte(caller), syntax.Options{Version: phpversion.Max})
	e, own := Summarize(use, resourceTypes(use, ix), snapshot, Options{})
	x := argument(callsNamed(use, "probe")[0])
	if !e.Tainted(x, HTML) || !e.Value(x).Complete {
		t.Fatal("published callable lost")
	}
	replacement := syntax.Parse("definitions.php", []byte(`<?php function factory(){return fn($x)=>htmlspecialchars($x);}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(replacement, ix, phpversion.Max, nil, Options{}))
	if own.Current(snapshot, next) {
		t.Fatal("returned body did not invalidate caller")
	}
}

func TestReturnedCallableRejectsInvalidBinding(t *testing.T) {
	e, file := summarized(t, `function factory(){return fn($x)=>htmlspecialchars($x);}$f=factory();probe($f(wrong:$_GET['x']));`)
	if e.Value(argument(callsNamed(file, "probe")[0])).Complete {
		t.Fatal("invalid named callback binding resolved")
	}
}

func TestReturnedCallableInvocationBudget(t *testing.T) {
	e, file := environment(t, `probe(1);`, nil)
	r := &scopeResult{complete: true, operations: MaxTransfers}
	if _, _, known := e.invokeReturned(callsNamed(file, "probe")[0], &callableValue{contract: &returnedCallable{}}, nil, nil, &frame{}, r); known || r.complete {
		t.Fatal("invocation budget ignored")
	}
}
