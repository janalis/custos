// Package splfixedarrayindexoutofbounds implements the native SplFixedArrayIndexOutOfBounds inspection.
package splfixedarrayindexoutofbounds

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an index within the fixed array size."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SplFixedArrayIndexOutOfBounds" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	f := n.(*syntax.ArrayDimFetch)
	size, known := semanticquery.NativeFixedArraySize(ctx, f.Var, f)
	i, ik := semanticquery.NativeInt(ctx, f.Dim)
	if known && ik && (i < 0 || i >= size) {
		ctx.ReportNode(f, message)
	}
}
