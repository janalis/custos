package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func contractSwitch(count int, expression, guard func(int) string) string {
	var src strings.Builder
	src.WriteString(`function f($choice,$input,$pattern){switch($choice){`)
	for i := range count {
		if i == count-1 {
			src.WriteString("default:")
		} else {
			fmt.Fprintf(&src, "case %d:", i)
		}
		fmt.Fprintf(&src, "$v=%s;", expression(i))
		if guard != nil {
			src.WriteString(guard(i))
		}
		src.WriteString("break;")
	}
	src.WriteString(`}probe($v);}`)
	return src.String()
}

func TestFiniteProducerContractQueries(t *testing.T) {
	for _, guarded := range []bool{false, true} {
		src := contractSwitch(70, func(i int) string {
			if i == 69 {
				return `preg_replace($pattern,'x',$input)`
			}
			return `file_get_contents($input)`
		}, func(i int) string {
			if guarded || i < 69 {
				return `if($v===false||$v===null){return;}`
			}
			return ""
		})
		probeNative(t, src, func(ctx *analysis.Context, at *syntax.FuncCall) {
			x := CallArgument(at.Args, 0, "")
			if got := NativeLocalBuiltinResult(ctx, x, "null", "preg_replace"); (got == nil) != guarded {
				t.Fatalf("guarded=%v preg=%T", guarded, got)
			}
			if NativeLocalBuiltinResult(ctx, x, "false", "file_get_contents") != nil {
				t.Fatal("false/null contracts or guard states conflated")
			}
			if NativeLocalResultOf(ctx, x, "null", func(syntax.Expr) bool { return true }) != nil || NativeLocalResult(ctx, x, "null") != nil || NativeLocalValue(ctx, x) != nil || NativeLocalOnly(ctx, x) != nil {
				t.Fatal("generic query accepted compressed identity")
			}
			if !NativeFailureFlowOnly(ctx, x) {
				t.Fatal("compressed producer paths lost flow-only status")
			}
		})
	}
	probeNative(t, contractSwitch(70, func(i int) string { return fmt.Sprintf("project%d($input)", i) }, nil), func(ctx *analysis.Context, at *syntax.FuncCall) {
		x := CallArgument(at.Args, 0, "")
		result := NativeLocalResultOf(ctx, x, "false", func(e syntax.Expr) bool {
			call, ok := e.(*syntax.FuncCall)
			return ok && call.Name.(*syntax.Name).Value == "project69"
		})
		if result == nil {
			t.Fatal("exact opaque origins lost arbitrary predicate")
		}
	})
	probeNative(t, contractSwitch(70, func(i int) string { return fmt.Sprint(i) }, nil), func(ctx *analysis.Context, at *syntax.FuncCall) {
		x := CallArgument(at.Args, 0, "")
		if !NativeFailureFlowOnly(ctx, x) || NativeLocalOnly(ctx, x) != nil || NativeLocalBuiltinResult(ctx, x, "false", "fopen") != nil {
			t.Fatal("syntax-free other paths lost ambiguity")
		}
	})
}

func TestFiniteProducerQueryHostBudget(t *testing.T) {
	probeNative(t, contractSwitch(1300, func(int) string { return `project($input)` }, nil), func(ctx *analysis.Context, at *syntax.FuncCall) {
		x := CallArgument(at.Args, 0, "")
		if !ctx.Flow().ProducerBudgetExceeded(x) || NativeLocalBuiltinResult(ctx, x, "false", "fopen") != nil || NativeLocalValue(ctx, x) != nil {
			t.Fatal("host budget produced a lexical witness")
		}
	})
}
