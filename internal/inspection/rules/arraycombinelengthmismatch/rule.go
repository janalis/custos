// Package arraycombinelengthmismatch implements the native ArrayCombineLengthMismatch inspection.
package arraycombinelengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply arrays with matching lengths."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayCombineLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_combine") {
		return
	}
	a, ak := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "keys"))
	b, bk := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 1, "values"))
	if ak && bk && len(a) != len(b) {
		ctx.ReportNode(call, message)
	}
}
