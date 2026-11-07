package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestParameterRead(t *testing.T) {
	for _, tc := range []struct {
		body        string
		read, known bool
	}{
		{`echo $x;`, true, true},
		{`$x = 1; echo $x;`, false, true},
		{`if ($flag) { $x = 1; } echo $x;`, true, true},
		{`if ($flag) { $x = 1; } else { $x = 2; } echo $x;`, false, true},
		{`while ($flag) { $x = 1; } echo $x;`, true, true},
		{`do { $x = 1; } while ($flag); echo $x;`, false, true},
		{`$x = $x + 1; echo $x;`, true, true},
		{`$x = unknown(); echo $x;`, false, true},
		{`if ($flag) { $x = unknown(); } echo $x;`, true, true},
		{`eval($code); echo $x;`, false, false},
		{`$saved = get_defined_vars(); $x = 1; echo $x;`, true, true},
		{`$saved = compact('x'); $x = 1; echo $x;`, true, true},
		{`$saved = func_get_args(); $x = 1; echo $x;`, true, true},
		{`$saved = func_get_arg(0); $x = 1; echo $x;`, true, true},
		{`$saved = $$name; $x = 1; echo $x;`, true, true},
		{`${$flag} = 1; $x = 1; echo $x;`, false, true},
	} {
		t.Run(tc.body, func(t *testing.T) {
			e, f := environment(t, `function f($x, $flag) { `+tc.body+` }`, nil)
			fn := f.Stmts[0].(*syntax.Function)
			read, known := e.ParameterRead(fn, 0)
			if read != tc.read || known != tc.known {
				t.Fatalf("got %v, %v; want %v, %v", read, known, tc.read, tc.known)
			}
			for _, index := range []int{-1, 2} {
				if _, ok := e.ParameterRead(fn, index); ok {
					t.Fatal("invalid parameter")
				}
			}
		})
	}
}

func BenchmarkParameterRead(b *testing.B) {
	e, f := environment(b, `function f($x, $flag) { if ($flag) { $x = 1; } else { $x = 2; } echo $x; }`, nil)
	fn := f.Stmts[0].(*syntax.Function)
	e.ParameterRead(fn, 0)
	b.ReportAllocs()
	for b.Loop() {
		e.ParameterRead(fn, 0)
	}
}

func TestParameterReadInvalidDeclarations(t *testing.T) {
	e, f := environment(t, `function f(&$x) { echo $x; }`, nil)
	if _, known := e.ParameterRead(f.Stmts[0], 0); known {
		t.Fatal("reference parameter")
	}
}

func TestParameterReadCaptures(t *testing.T) {
	for _, body := range []string{`$cb = fn() => $x; $x = 1;`, `$cb = function() use ($x) {}; $x = 1;`, `class Nested {} function other() {} $cb = function() use (&$x) {};`} {
		e, f := environment(t, `function f($x) { `+body+` }`, nil)
		if read, known := e.ParameterRead(f.Stmts[0], 0); !read && known {
			t.Fatal("capture proved unused")
		}
	}
}

func TestParameterObservationPoints(t *testing.T) {
	for _, tc := range []struct {
		body        string
		read, known bool
	}{
		{`$x=1; $cb=fn()=> $x; echo $x;`, false, true},
		{`$cb=fn($x)=> $x; $x=1; echo $x;`, false, true},
		{`$x=1; $cb=fn()=>fn()=> $x; echo $x;`, false, true},
		{`$cb=fn()=>fn()=> $x; $x=1; echo $x;`, true, true},
		{`$x=1; $saved=get_defined_vars(); echo $x;`, false, true},
		{`$saved=compact('flag'); $x=1; echo $x;`, false, true},
		{`$saved=func_get_arg(1); $x=1; echo $x;`, false, true},
		{`$x=1; $saved=func_get_args(); echo $x;`, false, true},
		{`$x=1; $saved=$$name; echo $x;`, false, true},
		{`if ($flag) { $x=1; } $saved=get_defined_vars(); echo $x;`, true, true},
		{`$x=unknown(); $saved=get_defined_vars(); echo $x;`, false, true},
		{`$x=1; $saved=compact(...$names); echo $x;`, false, true},
		{`$saved=compact(...$names); $x=1; echo $x;`, true, true},
		{`$saved=compact($name); $x=1; echo $x;`, true, true},
		{`$x=1; $cb=fn()=> $$name; echo $x;`, false, true},
		{`$x=1; $cb=fn()=>function()use($x){}; echo $x;`, false, true},
		{`$x=1; $cb=fn()=> C::$x; echo $x;`, false, true},
	} {
		e, f := environment(t, `function f($x,$flag){`+tc.body+`}`, nil)
		read, known := e.ParameterRead(f.Stmts[0], 0)
		if read != tc.read || known != tc.known {
			t.Fatalf("%s: %v,%v", tc.body, read, known)
		}
	}
}

func BenchmarkParameterObservations(b *testing.B) {
	e, f := environment(b, `function f($x,$flag){$x=1;$cb=fn()=>fn()=> $x;$saved=compact('flag');return [$x,$cb,$saved];}`, nil)
	b.ReportAllocs()
	for b.Loop() {
		New(f, e.types, nil, Options{}).ParameterRead(f.Stmts[0], 0)
	}
}

func TestMayHideEntry(t *testing.T) {
	cut := make([]Source, MaxSources)
	for _, tc := range []struct {
		value Value
		want  bool
	}{
		{Value{}, false},
		{Value{replaced: true, Complete: true}, true},
		{Value{Sources: cut}, true},
		{Value{Sources: cut, Complete: true}, false},
	} {
		if got := mayHideEntry(tc.value); got != tc.want {
			t.Fatalf("%+v: got %v", tc.value, got)
		}
	}
}

func TestObserveHiddenEntry(t *testing.T) {
	_, f := environment(t, `function f($x){}`, nil)
	r := &scopeResult{}
	observeParameter("x", Value{Complete: true, replaced: true}, f.Stmts[0], r)
	if r.parameterObserved["x"] != 1 {
		t.Fatal("replaced binding proved unread")
	}
}

func TestLegacyParameterObservation(t *testing.T) {
	e, f := environment(t, `function f($x){$x=1; $saved=func_get_args(); echo $x;}`, nil)
	e.types.PHP = phpversion.PHP53
	if read, known := e.ParameterRead(f.Stmts[0], 0); read || known {
		t.Fatal("legacy arguments proved unused")
	}
}

func TestArrowCaptureBudgetAndRecovery(t *testing.T) {
	if names := arrowCaptureNames(&syntax.ArrowFunction{}); !names.complete || len(names.names) != 0 {
		t.Fatal("empty recovered arrow")
	}
	e, f := environment(t, `function f($x){$cb=fn()=>[`+strings.Repeat(`$x,`, MaxTransfers)+`];$x=1;echo $x;}`, nil)
	if read, known := e.ParameterRead(f.Stmts[0], 0); !read && known {
		t.Fatal("exhausted capture scan proved unused")
	}
	_, source := environment(t, `function f($x){echo $x;}`, nil)
	fn := source.Stmts[0]
	r := &scopeResult{complete: true, operations: MaxTransfers}
	e.observeParameters(fn.(*syntax.Function).Body.Stmts[0], &frame{}, r)
	if r.complete || len(r.gaps) == 0 {
		t.Fatal("observation budget ignored")
	}
}

func TestArrowClassBodyIsNotCaptured(t *testing.T) {
	e, f := environment(t, `function f($x){$cb=fn()=>new class {public $x;};$x=1;echo $x;}`, nil)
	if read, known := e.ParameterRead(f.Stmts[0], 0); read || !known {
		t.Fatal("class member captured as local")
	}
}

func TestArrowAnonymousConstructorArguments(t *testing.T) {
	e, f := environment(t, `function f($x){$cb=fn()=>new class($x){function __construct($value){}};$x=1;echo $x;}`, nil)
	if read, known := e.ParameterRead(f.Stmts[0], 0); !read || !known {
		t.Fatal("anonymous constructor argument not captured")
	}
}

func TestArrowRecoveredCapture(t *testing.T) {
	arrow := &syntax.ArrowFunction{Expr: &syntax.Closure{Uses: []*syntax.ClosureUse{{}}}}
	if arrowCaptureNames(arrow).complete {
		t.Fatal("missing capture was treated as complete")
	}
}
