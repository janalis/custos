// Package inversetriginvalidrealdomain implements the native InverseTrigInvalidRealDomain inspection.
package inversetriginvalidrealdomain

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep the inverse trigonometric argument within minus one and one."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "InverseTrigInvalidRealDomain" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "asin" && name != "acos" {
		return
	}
	value, known := number(ctx, semanticquery.CallArgument(c.Args, 0, "num"))
	if known && (value < -1 || value > 1) {
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
