package implodeargumentsorder

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// implodeArgumentsOrder reports implode() calls passing the separator (a
// string literal) as the second argument.
type implodeArgumentsOrder struct{}

func (implodeArgumentsOrder) ID() string               { return "ImplodeArgumentsOrder" }
func (implodeArgumentsOrder) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (implodeArgumentsOrder) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name, ok := call.Name.(*syntax.Name)
	if !ok || call.Args == nil {
		return
	}
	if len(call.Args.Args) != 2 || !ctx.IsGlobalFunctionCall(call, "implode") { // D1, E1, E3
		return
	}
	first, ok1 := call.Args.Args[0].(*syntax.Arg)
	second, ok2 := call.Args.Args[1].(*syntax.Arg)
	if !ok1 || !ok2 || first.Value == nil || second.Value == nil {
		return
	}
	if !astquery.IsStringLiteral(second.Value) || astquery.IsStringLiteral(first.Value) { // D3, E2, E5
		return
	}
	span := call.Span()
	repl := name.Value + "(" + ctx.Text(second.Value) + ", " + ctx.Text(first.Value) + ")"
	ctx.Report(span, "Pass the separator as the first argument of implode().", diagnostic.Fix{
		Title: "Swap the arguments",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}
