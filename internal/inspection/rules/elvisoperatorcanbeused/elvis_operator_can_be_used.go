package elvisoperatorcanbeused

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// elvisOperatorCanBeUsed reports `$a ? $a : $b`, better written `$a ?: $b`.
type elvisOperatorCanBeUsed struct{}

func (elvisOperatorCanBeUsed) ID() string               { return "ElvisOperatorCanBeUsed" }
func (elvisOperatorCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (elvisOperatorCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Cond == nil || t.Then == nil || t.Else == nil { // D1, D2
		return
	}
	if t.Else.Span().Len() == 0 { // error recovery: missing else branch
		return
	}
	span := t.Span()
	if !astquery.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(t.Cond), syntax.UnwrapParens(t.Then)) { // D3
		return
	}
	// E4: the short ternary evaluates the condition once; with calls or
	// other side effects that drops the second evaluation.
	if flowquery.MayHaveSideEffects(t.Cond) {
		return
	}
	repl := ctx.Text(t.Cond) + " ?: " + ctx.Text(t.Else) // D4
	ctx.Report(span, "Use the short ternary: '"+repl+"'.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}
