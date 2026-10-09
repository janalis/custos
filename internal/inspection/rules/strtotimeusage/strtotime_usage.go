package strtotimeusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// strtotimeUsage reports strtotime('now') (use time()) and a redundant
// time() base timestamp argument.
type strtotimeUsage struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (strtotimeUsage) Semantic()                {}
func (strtotimeUsage) ID() string               { return "StrtotimeUsage" }
func (strtotimeUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (strtotimeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "strtotime") // D1
	if call == nil || !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "strtotime") {
		return
	}
	args := astquery.WrittenArguments(call.Args)
	if call.Args == nil || len(args) != len(call.Args.Args) || len(args) == 0 || len(args) > 2 {
		return
	}
	span := call.Span()
	switch len(args) {
	case 1: // D2
		c, _, ok := astquery.QuotedStringRaw(args[0].Value)
		if !ok || !strings.EqualFold(c, "now") {
			return
		}
		repl := strtotimeTimeCall(ctx, call)
		ctx.Report(span, "Call time() instead of parsing 'now'.", diagnostic.Fix{
			Title: "Replace with 'time()'",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
		})
	case 2: // D3
		inner, _ := semanticquery.GlobalCall(ctx, args[1].Value, "time")
		if inner == nil || !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, inner, "time") {
			return
		}
		repl := ctx.Text(call.Name) + "(" + ctx.Text(args[0].Value) + ")" // F2: callee kept as written
		ctx.Report(span, "The base timestamp already defaults to the current time; drop the time() argument.", diagnostic.Fix{
			Title: "Drop the time() argument",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
		})
	}
}

// strtotimeTimeCall is the F1 replacement: `\time()` when the original call
// was fully qualified or when an unqualified `time` would not reach the
// global function at this point (namespaced function or import); `time()`
// otherwise.
func strtotimeTimeCall(ctx *analysis.Context, call *syntax.FuncCall) string {
	return semanticquery.QualifiedBuiltinFor(ctx, "time", call) + "()"
}
