// Package sortcomparatorreturnsboolean implements the native SortComparatorReturnsBoolean inspection.
package sortcomparatorreturnsboolean

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Return a three-way integer comparison result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SortComparatorReturnsBoolean" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "usort", "uasort", "uksort":
	default:
		return
	}
	_, body := semanticquery.NativeCallback(ctx, semanticquery.CallArgument(call.Args, 1, "callback"))
	values, complete := semanticquery.NativeReturns(body)
	if !complete {
		return
	}
	for _, value := range values {
		b, ok := syntax.UnwrapParens(value).(*syntax.Binary)
		if !ok {
			return
		}
		switch b.Op.Kind {
		case syntax.TGreater, syntax.TLess, syntax.TIsGreaterOrEqual, syntax.TIsSmallerOrEqual:
		default:
			return
		}
	}
	var fixes []diagnostic.Fix
	if len(values) == 1 && ctx.PHP >= phpversion.PHP70 {
		b := syntax.UnwrapParens(values[0]).(*syntax.Binary)
		if (b.Op.Kind == syntax.TGreater || b.Op.Kind == syntax.TLess) && !flowquery.MayHaveSideEffects(b.Left) && !flowquery.MayHaveSideEffects(b.Right) {
			edits := []diagnostic.TextEdit{{Span: b.Op.Span, NewText: "<=>"}}
			if b.Op.Kind == syntax.TLess {
				edits = append(edits,
					diagnostic.TextEdit{Span: b.Left.Span(), NewText: ctx.Text(b.Right)},
					diagnostic.TextEdit{Span: b.Right.Span(), NewText: ctx.Text(b.Left)},
				)
			}
			fixes = append(fixes, diagnostic.Fix{Title: "Return an integer ordering", Edits: func() []diagnostic.TextEdit { return edits }})
		}
	}
	ctx.ReportNode(call, message, fixes...)
}
