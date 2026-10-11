package implodeargumentsorder

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// implodeArgumentsOrder reports implode() calls passing the separator (a
// string literal) as the second argument.
type implodeArgumentsOrder struct{}

const message = "Pass the separator as the first argument of implode()."

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
	if first.Name != nil || second.Name != nil || first.Unpack || second.Unpack { // E6
		return
	}
	if !astquery.IsStringLiteral(second.Value) || astquery.IsStringLiteral(first.Value) { // D3, E2, E5
		return
	}
	span := call.Span()
	if ctx.PHP < phpversion.PHP73 && (hasHeredoc(first.Value) || hasHeredoc(second.Value)) { // F2
		ctx.Report(span, message)
		return
	}
	repl := name.Value + "(" + ctx.Text(second.Value) + ", " + ctx.Text(first.Value) + ")"
	ctx.Report(span, message, diagnostic.Fix{
		Title: "Swap the arguments",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// hasHeredoc reports whether e contains a heredoc or nowdoc. Before PHP 7.3
// its closing marker must end its line, so an argument holding one cannot
// be moved in front of a `,` or `)`.
func hasHeredoc(e syntax.Expr) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Literal:
			found = found || strings.HasPrefix(x.Raw, "<<<")
		case *syntax.InterpolatedString:
			found = found || x.Heredoc
		}
		return !found
	})
	return found
}
