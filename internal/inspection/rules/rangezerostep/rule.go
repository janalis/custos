// Package rangezerostep implements the native RangeZeroStep inspection.
package rangezerostep

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonzero range step."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RangeZeroStep" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "range") {
		return
	}
	v, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 2, "step"))
	if k && v == 0 {
		ctx.ReportNode(c, message)
		return
	}
	vf, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 2, "step")).(*syntax.Literal)
	if ok && vf.LitKind == syntax.LitFloat {
		number, err := strconv.ParseFloat(vf.Raw, 64)
		if err == nil && number == 0 {
			ctx.ReportNode(c, message)
		}
	}
}
