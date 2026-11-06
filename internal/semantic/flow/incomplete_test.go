package flow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestIncompleteFlowReasons(t *testing.T) {
	for _, tc := range []struct{ body, reason string }{
		{`$x=1; strlen($x);`, ""},
		{`eval($code); strlen("value");`, "dynamic-code"},
		{`missingCall();`, "unresolved-dispatch"},
		{`include 'generated.php'; strlen("value");`, "dynamic-code"},
		{`yield 1; strlen("value");`, ""},
		{`try{goto done;}finally{1;}done:strlen("value");`, ""},
		{`while($code){break 3;} strlen("value");`, "unsupported-control-flow"},
	} {
		t.Run(tc.body, func(t *testing.T) {
			e, f := environment(t, `function example($code){`+tc.body+`}`, nil)
			e.Calls(f.Stmts[0])
			gaps := e.Gaps()
			if tc.reason == "" {
				if len(gaps) != 0 {
					t.Fatal(gaps)
				}
				return
			}
			if len(gaps) != 1 || gaps[0].Reason != tc.reason {
				t.Fatalf("gaps=%+v, want %s", gaps, tc.reason)
			}
			if !reflect.DeepEqual(gaps, e.Gaps()) {
				t.Fatal("gap ordering is not deterministic")
			}
		})
	}
}

func TestFlowBudgetIsExplicit(t *testing.T) {
	e, f := environment(t, `function example($flag){`+strings.Repeat(`if($flag){work();}`, 5000)+`}`, nil)
	e.Calls(f.Stmts[0])
	gaps := e.Gaps()
	if len(gaps) == 0 || gaps[0].Reason != "flow-budget" {
		t.Fatalf("gaps=%+v", gaps)
	}
}

func TestIncompleteStatusIsBounded(t *testing.T) {
	var source strings.Builder
	for i := range 130 {
		fmt.Fprintf(&source, "function f%d(){eval('generated();');}", i)
	}
	e, f := environment(t, source.String(), nil)
	for _, scope := range f.Stmts {
		e.Calls(scope)
	}
	gaps := e.Gaps()
	if len(gaps) != 128 || gaps[127].Reason != "status-limit" {
		t.Fatalf("status cap: %d %+v", len(gaps), gaps[len(gaps)-1])
	}
}

// A fluent builder chain is deep without being nested code: its links do not
// exhaust the expression depth bound, while nesting the same calls does.
func TestFluentChainDepth(t *testing.T) {
	const links = MaxDepth * 8 * 3
	builder := `final class B { function add($x) { return $this; } } `
	var chain strings.Builder
	chain.WriteString("$b")
	for i := range links {
		fmt.Fprintf(&chain, "->add(%d)", i)
	}
	nested := strings.Repeat("$b->add(", links) + "1" + strings.Repeat(")", links)
	for _, tc := range []struct {
		body   string
		budget bool
	}{
		{chain.String() + ";", false},
		{nested + ";", true},
	} {
		e, f := closedEnvironment(t, builder+`function example(B $b){`+tc.body+`}`)
		e.Calls(f.Stmts[1])
		if budget := len(e.Gaps()) > 0; budget != tc.budget {
			t.Errorf("links=%d nested=%v: gaps %+v", links, tc.budget, e.Gaps())
		}
	}
}

func BenchmarkFluentChain(b *testing.B) {
	for _, links := range []int{50, 100, 200} {
		b.Run(fmt.Sprint(links), func(b *testing.B) {
			var chain strings.Builder
			chain.WriteString(`final class B { function add($x) { return $this; } } function example(B $b){ $b`)
			for i := range links {
				fmt.Fprintf(&chain, "->add(%d)", i)
			}
			chain.WriteString(";}")
			e, f := closedEnvironment(b, chain.String())
			b.ReportAllocs()
			for b.Loop() {
				New(f, e.types, e.snapshot, Options{}).Calls(f.Stmts[1])
			}
		})
	}
}
