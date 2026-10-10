// Package validatedintegerzerorejected implements the native ValidatedIntegerZeroRejected inspection.
package validatedintegerzerorejected

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare integer validation failure strictly with false."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ValidatedIntegerZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "filter_var") {
		return
	}
	filter, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "filter"))
	if !known || filter != 257 {
		return
	}
	options := semanticquery.CallArgument(call.Args, 2, "options")
	if !semanticquery.NativeIntegerAllowsZero(ctx, options) {
		return
	}
	p, _ := astquery.ParentSkipParens(call)
	u, ok := p.(*syntax.Unary)
	if !ok || u.Op.Kind != syntax.TExclaim {
		return
	}
	ctx.ReportNode(call, message, diagnostic.Fix{Title: "Distinguish invalid integers from zero", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: u.Op.Span, NewText: "("}, {Span: syntax.Span{Start: u.Span().End, End: u.Span().End}, NewText: " === false)"}}
	}})
}
