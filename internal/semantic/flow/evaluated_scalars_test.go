package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestScalarStringEvaluationBounds(t *testing.T) {
	for _, tc := range []struct {
		left, right string
		work        int
	}{
		{strings.Repeat("a", 2048), strings.Repeat("b", 2049), 0},
		{"a", "b", MaxTransfers},
	} {
		left, right := &syntax.Variable{Name: "left"}, &syntax.Variable{Name: "right"}
		expr := &syntax.Binary{Left: left, Right: right, Op: syntax.TokenRef{Kind: syntax.TDot}}
		r := &scopeResult{operations: tc.work}
		a, b := constantScalar{k: 's', s: tc.left}, constantScalar{k: 's', s: tc.right}
		r.values.set(left, Value{Complete: true, literal: &a})
		r.values.set(right, Value{Complete: true, literal: &b})
		got := evaluatedScalar(expr, Value{Complete: true, Expr: expr}, r)
		if _, known := knownScalar(got); known {
			t.Fatal("oversized or over-budget scalar expression became known")
		}
		if r.operations != tc.work {
			t.Fatal("rejected concatenation consumed work")
		}
	}
}

func TestScalarArithmeticRefusals(t *testing.T) {
	for _, source := range []string{
		`$cb=fn($x,$y)=>$x+$y;probe($cb(9223372036854775807,1));`,
		`$cb=fn($x,$y)=>$x+$y;probe($cb('1',2));`,
		`$cb=fn($x,$y)=>$x/$y;probe($cb(1,0));`,
		`$cb=fn($x)=>-$x;probe($cb(-9223372036854775807-1));`,
		`$cb=fn($x)=>+$x;probe($cb('1'));`,
	} {
		e, f := summarized(t, source)
		if scalar, known := knownScalar(e.Value(argument(callsNamed(f, "probe")[0]))); known {
			t.Fatalf("unsupported arithmetic became %+v: %s", scalar, source)
		}
	}
}
