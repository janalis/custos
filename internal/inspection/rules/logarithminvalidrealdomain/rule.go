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
	value, known := number(ctx, semanticquery.CallArgument(c.Args, 0, "num"))
	bad := known && ((name == "log1p" && value <= -1) || (name != "log1p" && value <= 0))
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
