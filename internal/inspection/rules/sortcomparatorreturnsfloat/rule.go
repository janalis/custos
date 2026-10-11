// Package sortcomparatorreturnsfloat implements the native SortComparatorReturnsFloat inspection.
package sortcomparatorreturnsfloat

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Compare float values without subtracting them."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SortComparatorReturnsFloat" }
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
	callback := semanticquery.CallArgument(call.Args, 1, "callback")
	_, body := semanticquery.NativeCallback(ctx, callback)
	values, complete := semanticquery.NativeReturns(body)
	if !complete {
		return
	}
	for _, value := range values {
		b, ok := syntax.UnwrapParens(value).(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TMinus || !ctx.TypeOf(b.Left).OnlyOf("float") || !ctx.TypeOf(b.Right).OnlyOf("float") {
			return
		}
	}
	var fixes []diagnostic.Fix
	// A named function or a closure held in a variable may be called elsewhere
	// for its arithmetic result, so only an inline comparator is rewritten.
	if len(values) == 1 && ctx.PHP >= phpversion.PHP70 && inlineCallback(callback) {
		b := syntax.UnwrapParens(values[0]).(*syntax.Binary)
		fixes = append(fixes, diagnostic.Fix{Title: "Compare without float truncation", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: b.Op.Span, NewText: "<=>"}} }})
	}
	ctx.ReportNode(call, message, fixes...)
}

// inlineCallback reports whether the argument is itself an anonymous function,
// which nothing but the sort call can invoke.
func inlineCallback(e syntax.Expr) bool {
	switch syntax.UnwrapParens(e).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		return true
	}
	return false
}
