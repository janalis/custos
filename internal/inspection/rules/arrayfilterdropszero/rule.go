// Package arrayfilterdropszero implements the native ArrayFilterDropsZero inspection.
package arrayfilterdropszero

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve zero values with an explicit filter predicate."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFilterDropsZero" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_filter") {
		return
	}
	callback := semanticquery.CallArgument(call.Args, 1, "callback")
	if callback != nil && !ctx.TypeOf(callback).OnlyOf("null") {
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if !known {
		return
	}
	zero, other := false, false
	for _, value := range entries {
		if i, ok := semanticquery.NativeInt(ctx, value); ok && i == 0 {
			zero = true
		}
		if s, ok := semanticquery.NativeString(ctx, value); ok && s == "0" {
			zero = true
		}
		if truth, known := semanticquery.NativeTruth(ctx, value); known && truth {
			other = true
		}
	}
	if zero && other {
		ctx.ReportNode(call, message)
	}
}
