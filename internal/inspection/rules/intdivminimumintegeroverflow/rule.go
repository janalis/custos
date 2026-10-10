// Package intdivminimumintegeroverflow implements the native IntdivMinimumIntegerOverflow inspection.
package intdivminimumintegeroverflow

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Avoid dividing PHP_INT_MIN by negative one."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntdivMinimumIntegerOverflow" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "intdiv") {
		return
	}
	v, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "num2"))
	constant, ok := semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "num1")).(*syntax.ConstFetch)
	if k && v == -1 && ok && semanticquery.GlobalConstName(ctx, constant) == "PHP_INT_MIN" {
		ctx.ReportNode(c, message)
	}
}
