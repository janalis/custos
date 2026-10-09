package senselessternaryoperator

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// senselessTernaryOperator reports `$a === $b ? $a : $b` style ternaries
// that always yield the same operand.
type senselessTernaryOperator struct{}

func (senselessTernaryOperator) ID() string               { return "SenselessTernaryOperator" }
func (senselessTernaryOperator) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (senselessTernaryOperator) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Then == nil || t.Else == nil { // D1/E2
		return
	}
	cond, ok := syntax.UnwrapParens(t.Cond).(*syntax.Binary) // D2
	if !ok || cond.Left == nil || cond.Right == nil {
		return
	}
	var yes, no syntax.Expr // D3
	switch cond.Op.Kind {
	case syntax.TIsIdentical:
		yes, no = t.Then, t.Else
	case syntax.TIsNotIdentical:
		yes, no = t.Else, t.Then
	default:
		return
	}
	f := ctx.File
	eq := func(a, b syntax.Expr) bool { return astquery.EquivalentFoldNames(f, a, b) }
	if !(eq(cond.Left, yes) && eq(cond.Right, no)) && !(eq(cond.Left, no) && eq(cond.Right, yes)) { // D4
		return
	}
	repl := ctx.Text(no) // D6
	span := t.Span()
	ctx.Report(span, "This ternary always yields '"+repl+"'; use it directly.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}
