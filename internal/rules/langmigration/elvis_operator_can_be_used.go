package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// elvisOperatorCanBeUsed reports `$a ? $a : $b`, better written `$a ?: $b`.
type elvisOperatorCanBeUsed struct{}

func init() { register(elvisOperatorCanBeUsed{}) }

func (elvisOperatorCanBeUsed) ID() string { return "ElvisOperatorCanBeUsed" }

func (elvisOperatorCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (elvisOperatorCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Cond == nil || t.Then == nil || t.Else == nil { // D1, D2
		return
	}
	span := t.Span()
	if span.Len() == 0 || t.Else.Span().Len() == 0 {
		return
	}
	if !util.EquivalentFoldNames(ctx.File, util.UnwrapParens(t.Cond), util.UnwrapParens(t.Then)) { // D3
		return
	}
	// E4: the short ternary evaluates the condition once; with calls or
	// other side effects that drops the second evaluation.
	if util.MayHaveSideEffects(t.Cond) {
		return
	}
	repl := ctx.Text(t.Cond) + " ?: " + ctx.Text(t.Else) // D4
	ctx.Report(span, "Use the short ternary: '"+repl+"'.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
