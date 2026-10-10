// Package weakmapvalueretainskey implements the native WeakMapValueRetainsKey inspection.
package weakmapvalueretainskey

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Store a value that does not retain the WeakMap key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WeakMapValueRetainsKey" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	a := n.(*syntax.Assign)
	slot, ok := a.Var.(*syntax.ArrayDimFetch)
	if !ok || semanticquery.NativeConstruction(ctx, slot.Var, "WeakMap") == nil {
		return
	}
	retained := semanticquery.NativeSameValue(ctx, slot.Dim, a.Value)
	if arr := semanticquery.NativeArray(ctx, a.Value); arr != nil {
		for _, item := range arr.Items {
			if item != nil && semanticquery.NativeSameValue(ctx, slot.Dim, item.Value) {
				retained = true
			}
		}
	}
	if retained {
		ctx.ReportNode(a, message)
	}
}
