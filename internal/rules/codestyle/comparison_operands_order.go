package codestyle

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// comparisonOperandsOrder enforces the configured operand order (yoda or
// regular) of equality comparisons against a literal or constant.
type comparisonOperandsOrder struct{}

func init() { register(comparisonOperandsOrder{}) }

func (comparisonOperandsOrder) ID() string { return "ComparisonOperandsOrder" }

func (comparisonOperandsOrder) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }

func (comparisonOperandsOrder) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	switch b.Op.Kind { // D1
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
	default:
		return // E4
	}
	if b.Left == nil || b.Right == nil || b.Left.Span().Len() == 0 || b.Right.Span().Len() == 0 {
		return
	}
	left, right := cooConstantLike(b.Left), cooConstantLike(b.Right)
	if left == right { // E1, E2
		return
	}
	msg := "Put the constant operand on the right side of the comparison."
	bad := left
	if ctx.ComparisonStyle == analysis.StyleYoda {
		msg = "Put the constant operand on the left side of the comparison."
		bad = right
	}
	if !bad { // E3
		return
	}
	ls, rs := b.Left.Span(), b.Right.Span()
	lt, rt := cooOperandText(ctx, b.Left), cooOperandText(ctx, b.Right)
	ctx.Report(syntax.Span{Start: ls.Start, End: rs.End}, msg, analysis.Fix{
		Title: "Swap the operands",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: ls, NewText: rt}, {Span: rs, NewText: lt}}
		},
	})
}

// cooConstantLike implements D2 (no parentheses unwrapping).
func cooConstantLike(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Literal, *syntax.ConstFetch, *syntax.MagicConst:
		return true
	case *syntax.InterpolatedString:
		return !e.Backtick
	case *syntax.Unary:
		if e.Op.Kind != syntax.TMinus {
			return false
		}
		l, ok := e.Expr.(*syntax.Literal)
		return ok && (l.LitKind == syntax.LitInt || l.LitKind == syntax.LitFloat)
	}
	return false
}

// cooOperandText is the operand's text for the other side: an operand
// binding looser than the comparison is parenthesized, or
// `false !== $r = f()` would become `$r = f() !== false`.
func cooOperandText(ctx *analysis.Context, e syntax.Expr) string {
	if util.NeedsParensAsEqualityOperand(e) {
		return "(" + ctx.Text(e) + ")"
	}
	return ctx.Text(e)
}
