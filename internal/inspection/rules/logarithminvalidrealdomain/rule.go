// Package logarithminvalidrealdomain implements the native LogarithmInvalidRealDomain inspection.
package logarithminvalidrealdomain

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a positive logarithm argument."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LogarithmInvalidRealDomain" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "log" && name != "log10" && name != "log1p" {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "num")
	value, known := number(ctx, arg)
	bad := known && ((name == "log1p" && value <= -1) || (name != "log1p" && value <= 0))
	// E3: `log(0)` and `log1p(-1)` written out are the way to get -INF.
	if bad && literalBoundary(arg, name) {
		bad = false
	}
	if name == "log" {
		base, known := number(ctx, semanticquery.CallArgument(c.Args, 1, "base"))
		bad = bad || (known && (base <= 0 || base == 1))
	}
	if bad {
		ctx.ReportNode(c, message)
	}
}

func number(ctx *analysis.Context, e syntax.Expr) (float64, bool) {
	if n, ok := semanticquery.NativeInt(ctx, e); ok {
		return float64(n), true
	}
	value := semanticquery.NativeValue(ctx, e)
	sign := 1.0
	if u, ok := value.(*syntax.Unary); ok {
		if u.Op.Kind != syntax.TMinus && u.Op.Kind != syntax.TPlus {
			return 0, false
		}
		if u.Op.Kind == syntax.TMinus {
			sign = -1
		}
		value = syntax.UnwrapParens(u.Expr)
	}
	lit, ok := value.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return 0, false
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	return sign * n, err == nil
}

// literalBoundary reports whether e is written as a number literal equal to
// the edge of the domain of name (0, or -1 for log1p), whose logarithm is
// -INF without any warning.
func literalBoundary(e syntax.Expr, name string) bool {
	x := syntax.UnwrapParens(e)
	neg := false
	if u, ok := x.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		neg, x = true, syntax.UnwrapParens(u.Expr)
	}
	lit, ok := x.(*syntax.Literal)
	if !ok || lit.LitKind == syntax.LitString {
		return false
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	if err != nil {
		return false
	}
	if neg {
		n = -n
	}
	if name == "log1p" {
		return n == -1
	}
	return n == 0
}
