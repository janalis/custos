package performance

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// strtotimeUsage reports strtotime('now') (use time()) and a redundant
// time() base timestamp argument.
type strtotimeUsage struct{}

func init() { register(strtotimeUsage{}) }

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (strtotimeUsage) Semantic() {}

func (strtotimeUsage) ID() string { return "StrtotimeUsage" }

func (strtotimeUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (strtotimeUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := perfCall(ctx, n, "strtotime") // D1
	if call == nil || !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "strtotime") {
		return
	}
	args := perfArgs(call.Args)
	if call.Args == nil || len(args) != len(call.Args.Args) || len(args) == 0 || len(args) > 2 {
		return
	}
	span := call.Span()
	switch len(args) {
	case 1: // D2
		c, _, ok := util.QuotedStringRaw(args[0].Value)
		if !ok || !strings.EqualFold(c, "now") {
			return
		}
		repl := strtotimeTimeCall(ctx, call)
		ctx.Report(span, "Call time() instead of parsing 'now'.", analysis.Fix{
			Title: "Replace with 'time()'",
			Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
		})
	case 2: // D3
		inner, _ := perfCall(ctx, args[1].Value, "time")
		if inner == nil || !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, inner, "time") {
			return
		}
		repl := ctx.Text(call.Name) + "(" + ctx.Text(args[0].Value) + ")" // F2: callee kept as written
		ctx.Report(span, "The base timestamp already defaults to the current time; drop the time() argument.", analysis.Fix{
			Title: "Drop the time() argument",
			Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
		})
	}
}

// strtotimeTimeCall is the F1 replacement: `\time()` when the original call
// was fully qualified or when an unqualified `time` would not reach the
// global function at this point (namespaced function or import); `time()`
// otherwise.
func strtotimeTimeCall(ctx *analysis.Context, call *syntax.FuncCall) string {
	return util.QualifiedBuiltinFor(ctx, "time", call) + "()"
}
