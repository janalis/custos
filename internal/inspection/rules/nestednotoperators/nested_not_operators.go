package nestednotoperators

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// nestedNotOperators reports chains of two or more `!` (possibly through
// parentheses) and suggests a `(bool)` cast or a single `!`.
type nestedNotOperators struct{}

func (nestedNotOperators) ID() string               { return "NestedNotOperators" }
func (nestedNotOperators) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func isNot(n syntax.Node) (*syntax.Unary, bool) {
	u, ok := n.(*syntax.Unary)
	return u, ok && u.Op.Kind == syntax.TExclaim
}

func (nestedNotOperators) Check(ctx *analysis.Context, n syntax.Node) {
	inner, ok := isNot(n)
	if !ok || inner.Expr == nil {
		return
	}
	x := syntax.UnwrapParens(inner.Expr)
	if _, ok := isNot(x); ok {
		return // D1: only the innermost `!` starts a chain
	}
	// D2/D3: walk outwards through parentheses and `!` only (any other
	// unary operator ends the chain; see spec Divergences).
	count, outer := 1, inner
	for {
		parent, child := astquery.ParentSkipParens(outer)
		u, ok := isNot(parent)
		if !ok || u.Expr != child {
			break
		}
		count++
		outer = u
	}
	if count < 2 {
		return // E1
	}
	// D4/D5
	subject := ctx.Text(x)
	if astquery.NeedsParensAsUnaryOperand(x) {
		subject = "(" + subject + ")"
	}
	repl := "!" + subject
	if count%2 == 0 {
		repl = "(bool)" + subject
	}
	span := outer.Span()
	ctx.Report(span, "Simplify the stacked negations to '"+repl+"'.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}
