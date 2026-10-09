package dateusage

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// dateUsage reports `date($format, time())`: the timestamp argument is the
// default already.
type dateUsage struct{}

func (dateUsage) ID() string               { return "DateUsage" }
func (dateUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (dateUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !ctx.IsGlobalFunctionCall(call, "date") || call.Args == nil || len(call.Args.Args) != 2 { // D1/D2
		return
	}
	second, ok := call.Args.Args[1].(*syntax.Arg)
	if !ok || second.Name != nil || second.Unpack {
		return
	}
	inner, ok := second.Value.(*syntax.FuncCall)
	if !ok || !ctx.IsGlobalFunctionCall(inner, "time") || inner.Args == nil || len(inner.Args.Args) != 0 { // D3/D4
		return
	}
	del := syntax.Span{Start: call.Args.Args[0].Span().End, End: second.Span().End}
	ctx.ReportNode(inner, "Redundant time() argument: date() uses the current time by default.", diagnostic.Fix{
		Title: "Drop the time() argument",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: del}} },
	})
}
