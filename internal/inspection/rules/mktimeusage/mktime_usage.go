package mktimeusage

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// mktimeUsage reports argument-less mktime()/gmmktime() (use time()) and the
// removed is_dst seventh argument.
type mktimeUsage struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (mktimeUsage) Semantic()                {}
func (mktimeUsage) ID() string               { return "MktimeUsage" }
func (mktimeUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (mktimeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := ctx.GlobalFunctionName(call) // D1, D2 (any case; not a namespaced same-named function)
	if name != "mktime" && name != "gmmktime" {
		return
	}
	switch astquery.ArgCount(call) {
	case 0: // D4
		span := call.Span()
		repl := semanticquery.QualifiedBuiltin(ctx, "time", call.Span().Start) + "()" // a namespaced time() would capture a bare call
		ctx.ReportSeverity(span, diagnostic.SeverityWarning, "Call time() instead; mktime()/gmmktime() without arguments is deprecated.", diagnostic.Fix{
			Title: "Replace with '" + repl + "'",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
		})
	case 7: // D5
		if a, ok := call.Args.Args[6].(*syntax.Arg); ok && a.Value != nil {
			ctx.ReportSeverity(a.Value.Span(), diagnostic.SeverityWarning, "The is_dst argument is deprecated and was removed in PHP 7.0.")
		}
	}
}
