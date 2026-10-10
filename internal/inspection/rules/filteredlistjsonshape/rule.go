// Package filteredlistjsonshape implements the native FilteredListJsonShape inspection.
package filteredlistjsonshape

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reindex the filtered array before encoding a JSON list."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FilteredListJsonShape" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	forced, known := semanticquery.NativeFlag(ctx, flags, 16) // JSON_FORCE_OBJECT
	if !known || forced {
		return
	}
	e := semanticquery.CallArgument(call.Args, 0, "value")
	inner, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, inner, "array_filter") {
		return
	}
	callback := semanticquery.CallArgument(inner.Args, 1, "callback")
	if callback != nil && !ctx.TypeOf(callback).OnlyOf("null") {
		return
	}
	a := semanticquery.NativeArray(ctx, semanticquery.CallArgument(inner.Args, 0, "array"))
	if a == nil {
		return
	}
	gap, broken := false, false
	for i, item := range a.Items {
		if item.Key != nil {
			key, known := semanticquery.NativeInt(ctx, item.Key)
			if !known || key != int64(i) {
				return
			}
		}
		truth, known := semanticquery.NativeTruth(ctx, item.Value)
		if !known {
			return
		}
		if !truth {
			gap = true
		} else if gap {
			broken = true
		}
	}
	if broken {
		ctx.ReportNode(call, message, diagnostic.Fix{Title: "Reindex the filtered list", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: inner.Span(), NewText: "\\array_values(" + ctx.Text(inner) + ")"}}
		}})
	}
}
