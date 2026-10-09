package suspiciousfunctioncalls

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// suspiciousFunctionCalls reports string comparison calls whose two compared
// operands are the same expression.
type suspiciousFunctionCalls struct{}

const suspiciousFunctionCallsMsg = "Both compared strings are the same expression; one of them is probably wrong."

func (suspiciousFunctionCalls) ID() string { return "SuspiciousFunctionCalls" }
func (suspiciousFunctionCalls) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (suspiciousFunctionCalls) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch ctx.GlobalFunctionName(call) { // D1 (case-insensitive, global function only)
	case "strcmp", "strncmp", "strcasecmp", "strncasecmp", "strnatcmp",
		"strnatcasecmp", "substr_compare", "hash_equals":
	default:
		return
	}
	if call.Args == nil || len(call.Args.Args) < 2 { // D2
		return
	}
	a, ok1 := call.Args.Args[0].(*syntax.Arg)
	b, ok2 := call.Args.Args[1].(*syntax.Arg)
	if !ok1 || !ok2 || a.Value == nil || b.Value == nil || a.Unpack || b.Unpack {
		return
	}
	if astquery.EquivalentFoldNames(ctx.File, a.Value, b.Value) { // D3
		ctx.ReportNode(call, suspiciousFunctionCallsMsg)
	}
}
