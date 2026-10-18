package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestReferenceResult(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		known, taint bool
	}{
		{"literal replacement", `function fill(&$to,$from){$to=['v'=>'safe'];} fill($_GET,[]);`, true, false},
		{"named replacement", `function fill(&$to,$from){$to=$from;} fill(from:['v'=>'safe'],to:$_GET);`, true, false},
		{"retained argument", `function fill(&$to,$from){} fill($_GET,[]);`, true, true},
		{"conditional replacement", `function fill(&$to,$from){if($from){$to=[];}} fill($_GET,$c);`, true, true},
		{"copied request", `function fill(&$to,$from){$to=$from;} fill(from:$_POST,to:$_GET);`, true, true},
		{"by value", `function fill($to){$to=[];} fill($_GET);`, false, false},
		{"indirect global mutation", `function fill(&$to){$to=[];other();} fill($_GET);`, false, false},
		{"unresolved", `fill($_GET);`, false, false},
		{"conditional declaration", `if($c){function fill(&$to){$to=[];}} fill($_GET);`, false, false},
		{"variadic", `function fill(&...$to){} fill($_GET);`, false, false},
		{"spread", `function fill(&$to,$from){$to=[];} fill($_GET,...$args);`, false, false},
		{"multiple references", `function fill(&$to,&$from){$to=[];$from=$_POST;} fill($_GET,$_GET);`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := summarized(t, tc.source)
			call := callsNamed(f, "fill")[0]
			arg := call.Args.Args[0].(*syntax.Arg)
			// The named replacement's target is written second.
			if tc.name == "named replacement" || tc.name == "copied request" {
				arg = call.Args.Args[1].(*syntax.Arg)
			}
			v, known := e.ReferenceResult(call, arg)
			if known != tc.known || known && tainted(v) != tc.taint {
				t.Fatalf("known=%v value=%+v", known, v)
			}
		})
	}
}

func TestReferenceResultUnavailable(t *testing.T) {
	e, f := summarized(t, `function fill(&$to){$to=[];} fill($_GET);`)
	call := callsNamed(f, "fill")[0]
	arg := call.Args.Args[0].(*syntax.Arg)
	for _, invalid := range []*syntax.FuncCall{nil, {}, {Args: &syntax.ArgList{}}} {
		if _, known := e.ReferenceResult(invalid, arg); known {
			t.Fatal("invented result for absent call")
		}
	}
	if _, known := e.ReferenceResult(call, nil); known {
		t.Fatal("invented result for absent argument")
	}
	if _, known := e.ReferenceResult(call, &syntax.Arg{}); known {
		t.Fatal("invented result for foreign argument")
	}
	r := e.scope(nil)
	saved := r.values[arg.Value]
	delete(r.values, arg.Value)
	if _, known := e.ReferenceResult(call, arg); known {
		t.Fatal("invented result for unevaluated argument")
	}
	r.values[arg.Value] = saved
	r.complete = false
	if _, known := e.ReferenceResult(call, arg); known {
		t.Fatal("used incomplete scope")
	}
	r.complete = true
	_, name, _ := e.declaration(call)
	sum, _ := e.contract(name)
	delete(sum.ByRef, 0)
	if _, known := e.ReferenceResult(call, arg); known {
		t.Fatal("invented undocumented reference output")
	}
}

func TestReferenceResultBudget(t *testing.T) {
	var params strings.Builder
	for i := range 64 {
		fmt.Fprintf(&params, ",$p%d=null", i)
	}
	for _, source := range []string{
		`function fill(&$to){} fill($_GET` + strings.Repeat(`,0`, 256) + `);`,
		`function fill(&$to` + params.String() + `){} fill($_GET);`,
	} {
		e, f := summarized(t, source)
		call := callsNamed(f, "fill")[0]
		if _, known := e.ReferenceResult(call, call.Args.Args[0].(*syntax.Arg)); known {
			t.Fatal("budget exceeded but result proven")
		}
		found := false
		for _, gap := range e.Gaps() {
			found = found || gap.Reason == "argument-budget"
		}
		if !found {
			t.Fatal("missing argument-budget status")
		}
	}
}

func BenchmarkReferenceResult(b *testing.B) {
	e, f := summarized(b, `function fill(&$to,$from){$to=$from;} fill(from:['v'=>'safe'],to:$_GET);`)
	call := callsNamed(f, "fill")[0]
	arg := call.Args.Args[1].(*syntax.Arg)
	e.ReferenceResult(call, arg)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.ReferenceResult(call, arg)
	}
}
