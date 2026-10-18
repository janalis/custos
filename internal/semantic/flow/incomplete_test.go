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
