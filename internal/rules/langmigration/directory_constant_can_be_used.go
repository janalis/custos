package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// directoryConstantCanBeUsed reports dirname(__FILE__), which is __DIR__.
type directoryConstantCanBeUsed struct{}

func init() { register(directoryConstantCanBeUsed{}) }

func (directoryConstantCanBeUsed) ID() string { return "DirectoryConstantCanBeUsed" }

func (directoryConstantCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (directoryConstantCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	qual, _, ok := util.CallName(call)
	// D1 (any case, resolving to the global function); only unqualified or
	// `\`-qualified calls (spec Divergences).
	if !ok || (qual != "" && qual != `\`) || ctx.GlobalFunctionName(call) != "dirname" {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	mc, ok := args[0].(*syntax.MagicConst)    // D3
	if !ok || mc.Token.Kind != syntax.TFile { // any case: __file__ too
		return
	}
	span := call.Span()
	ctx.Report(span, "Replace dirname(__FILE__) with __DIR__.", analysis.Fix{
		Title: "Replace with __DIR__",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: "__DIR__"}} },
	})
}
