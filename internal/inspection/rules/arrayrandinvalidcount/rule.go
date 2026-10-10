// Package arrayrandinvalidcount implements the native ArrayRandInvalidCount inspection.
package arrayrandinvalidcount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Request a count within the array size."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayRandInvalidCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "array_rand") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 1, "num")
	count, k := int64(1), true
	if arg != nil {
		count, k = semanticquery.NativeInt(ctx, arg)
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 0, "array"))
	if k && (count < 1 || known && count > int64(len(entries))) {
		ctx.ReportNode(c, message)
	}
}
