package infer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestInferenceGapsFromSource(t *testing.T) {
	var parameters, arguments []string
	for i := range 65 {
		parameters = append(parameters, fmt.Sprintf("$p%d", i))
		arguments = append(arguments, "1")
	}
	template := `/** @template T
 * @param T $p0
 * @return T
 */ function identity(` + strings.Join(parameters, ",") + `){return $p0;}`
	var chain strings.Builder
	for i := range 7 {
		fmt.Fprintf(&chain, "function f%d(){return ", i)
		if i == 6 {
			chain.WriteString("1;")
		} else {
			fmt.Fprintf(&chain, "f%d();", i+1)
		}
		chain.WriteString("}")
	}
	for _, tc := range []struct{ source, reason string }{
		{template + `probe(identity(` + strings.Join(arguments, ",") + `));`, "argument-budget"},
		{`probe(max(` + strings.Repeat("1,", 256) + `1));`, "argument-budget"},
		{chain.String() + `probe(f0());`, "inference-budget"},
		{`function cycle(){return cycle();}probe(cycle());`, "inference-recursion"},
		{`probe(strlen('text'));`, ""},
	} {
		env, file := assertionEnv(t, tc.source, false, phpversion.PHP84)
		env.TypeOf(assertionRead(file))
		gaps := env.Gaps()
		if tc.reason == "" {
			if len(gaps) != 0 {
				t.Fatal(gaps)
			}
			continue
		}
		found := false
		for _, gap := range gaps {
			found = found || gap.Reason == tc.reason
		}
		if !found {
			t.Fatalf("missing %s: %+v", tc.reason, gaps)
		}
	}
}

func TestInferenceGapOwnershipAndBounds(t *testing.T) {
	env, _ := assertionEnv(t, `probe(1);`, false, phpversion.PHP84)
	native := env.Native()
	configured := native.WithOptions(Options{Assertions: true})
	for i := 1000; i >= 1; i-- {
		configured.noteGap(gapInferenceBudget, syntax.Span{Start: uint32(i), End: uint32(i + 4)})
	}
	native.noteGap(gapRecursion, syntax.Span{Start: 1, End: 2})
	configured.Native().noteGap(gapArgumentBudget, syntax.Span{Start: 1, End: 2})
	configured.noteGap(gapInferenceBudget, syntax.Span{Start: 1, End: 3})
	if len(env.gaps) != 3 || !reflect.DeepEqual(env.Gaps(), configured.Gaps()) {
		t.Fatal(env.Gaps(), configured.Gaps())
	}
	got := env.Gaps()
	if got[0].Reason != "argument-budget" || got[1].Reason != "inference-recursion" || got[2].Span.End != 3 {
		t.Fatal(got)
	}
	env.noteGap(gapArgumentBudget, syntax.Span{Start: 0, End: 1})
	if env.Gaps()[0].Span.Start != 0 {
		t.Fatal(env.Gaps())
	}
	if _, state := env.boundArgument(nil, make([]index.Param, 65), 0); state != bindingUnknown {
		t.Fatal(state)
	}
}

func TestCaptureAndExpressionGapBudgets(t *testing.T) {
	env, file := assertionEnv(t, `probe(fn()=>$value);`, false, phpversion.PHP84)
	closure := assertionRead(file)
	env.captureDepth = maxCaptureDepth
	env.captureType(closure, "value")
	env.captureDepth = 0
	env.captureBusy = map[captureKey]bool{{closure, "value"}: true}
	env.captureType(closure, "value")
	env.busy[closure] = true
	env.TypeOf(closure)
	if len(env.Gaps()) != 2 {
		t.Fatal(env.Gaps())
	}
	env.captureBusy = nil
	// Enough independent writes exceed the capture's reaching-definition budget.
	env, file = assertionEnv(t, `function sample(){`+strings.Repeat(`$value=1;`, maxVarDefs+1)+`probe(fn()=>$value);}`, false, phpversion.PHP84)
	closure = assertionRead(file)
	env.captureType(closure, "value")
	if len(env.Gaps()) != 1 || env.Gaps()[0].Reason != "inference-budget" {
		t.Fatal(env.Gaps())
	}
}

func BenchmarkInferenceGapCollection(b *testing.B) {
	env := &Env{}
	b.ReportAllocs()
	for b.Loop() {
		_ = env.Gaps()
	}
}

func BenchmarkTrackedArgumentBinding(b *testing.B) {
	env := &Env{}
	value := &syntax.Literal{LitKind: syntax.LitInt, Raw: "1"}
	argument := &syntax.Arg{Value: value}
	args := &syntax.ArgList{Args: []syntax.Expr{argument}}
	params := []index.Param{{Name: "value"}}
	b.ReportAllocs()
	for b.Loop() {
		env.boundArgument(args, params, 0)
	}
}
