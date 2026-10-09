package incorrectrandomrange

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// incorrectRandomRange reports rand()/mt_rand()/random_int() calls whose
// known bounds are in the wrong order.
type incorrectRandomRange struct{}

func (incorrectRandomRange) ID() string               { return "IncorrectRandomRange" }
func (incorrectRandomRange) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (incorrectRandomRange) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch ctx.GlobalFunctionName(call) { // D1
	case "rand", "mt_rand", "random_int":
	default:
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 2 { // D2
		return
	}
	from, ok := randomBound(ctx, args[0]) // D3
	if !ok {
		return
	}
	to, ok := randomBound(ctx, args[1]) // D4
	if !ok {
		return
	}
	if to < from { // D6
		ctx.ReportNode(call, "Minimum is greater than maximum in this random range.")
	}
}

// randomBound discovers the single numeric value of e and parses it as a
// signed decimal integer (D3-D6).
func randomBound(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	vals := semanticquery.DiscoverValues(ctx.Types(), e)
	if len(vals) != 1 {
		return 0, false
	}
	switch v := vals[0].(type) {
	case *syntax.Literal, *syntax.Unary:
		return numberNodeValue(v)
	case *syntax.ConstFetch: // constant declared outside the file (stubs)
		if k := semanticquery.ResolveConstant(ctx.Types(), v); k != nil {
			return parseDecimalBound(strings.TrimSpace(k.Value))
		}
	}
	return 0, false
}

// numberNodeValue returns the integer PHP passes for a number node (D5,
// D6): integer literals in any base (with `_` separators), floats truncated
// toward zero, optionally negated.
func numberNodeValue(e syntax.Expr) (int64, bool) {
	neg := false
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		neg, e = true, u.Expr
	}
	lit, ok := e.(*syntax.Literal)
	if !ok {
		return 0, false
	}
	var v int64
	switch lit.LitKind {
	case syntax.LitInt:
		if v, ok = astquery.ParseIntLiteral(lit.Raw); !ok {
			return 0, false
		}
	case syntax.LitFloat:
		f, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
		if err != nil || math.IsInf(f, 0) || math.IsNaN(f) || math.Abs(f) >= 1<<62 {
			return 0, false
		}
		v = int64(f) // truncation toward zero, as PHP's int conversion
	default:
		return 0, false
	}
	if neg {
		v = -v
	}
	return v, true
}

func parseDecimalBound(s string) (int64, bool) {
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err == nil
}

// Semantic marks the rule as needing the project index.
func (incorrectRandomRange) Semantic() {}
