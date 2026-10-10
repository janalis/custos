// Package arrayfillnegativecount implements the native ArrayFillNegativeCount inspection.
package arrayfillnegativecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonnegative array fill count."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFillNegativeCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "array_fill") {
		return
	}
	v, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "count"))
	if k && v < 0 {
		ctx.ReportNode(c, message)
	}
}
