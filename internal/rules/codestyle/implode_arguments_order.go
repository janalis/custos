package codestyle

import (
	"custos/internal/analysis"
	"custos/internal/syntax"
)

// implodeArgumentsOrder reports implode() calls passing the separator (a
// string literal) as the second argument.
type implodeArgumentsOrder struct{}

func init() { register(implodeArgumentsOrder{}) }

func (implodeArgumentsOrder) ID() string { return "ImplodeArgumentsOrder" }

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
	if !iaoStringLiteral(second.Value) || iaoStringLiteral(first.Value) { // D3, E2, E5
		return
	}
	span := call.Span()
	repl := name.Value + "(" + ctx.Text(second.Value) + ", " + ctx.Text(first.Value) + ")"
	ctx.Report(span, "Pass the separator as the first argument of implode().", analysis.Fix{
		Title: "Swap the arguments",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

func iaoStringLiteral(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Literal:
		return e.LitKind == syntax.LitString
	case *syntax.InterpolatedString:
		return !e.Backtick
	}
	return false
}
