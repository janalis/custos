package comparisonoperandsorder

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// comparisonOperandsOrder enforces the configured operand order (yoda or
// regular) of equality comparisons against a literal or constant.
type comparisonOperandsOrder struct{}

func (comparisonOperandsOrder) ID() string               { return "ComparisonOperandsOrder" }
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
	span := syntax.Span{Start: ls.Start, End: rs.End}
	if !cooSwappable(ctx, b) { // F2, F3
		ctx.Report(span, msg)
		return
	}
	lt, rt := cooOperandText(ctx, b.Left), cooOperandText(ctx, b.Right)
	// F4: keep the moved operands apart from a word glued to the comparison.
	src := ctx.File.Src
	if ls.Start > 0 && cooWordByte(src[ls.Start-1]) && (cooWordByte(rt[0]) || rt[0] == '\\') {
		rt = " " + rt
	}
	if int(rs.End) < len(src) && cooWordByte(src[rs.End]) && cooWordByte(lt[len(lt)-1]) {
		lt += " "
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Swap the operands",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: ls, NewText: rt}, {Span: rs, NewText: lt}}
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

// cooSwappable reports whether exchanging the operands keeps the program's
// behaviour and syntax: an interpolated string reads its variables when it
// is evaluated, so it must not change places with an operand that has side
// effects (F2), and before PHP 7.3 a heredoc closing marker has to stay
// alone on its line (F3).
func cooSwappable(ctx *analysis.Context, b *syntax.Binary) bool {
	for _, e := range []syntax.Expr{b.Left, b.Right} {
		if _, ok := e.(*syntax.InterpolatedString); ok &&
			(flowquery.MayHaveSideEffects(b.Left) || flowquery.MayHaveSideEffects(b.Right)) {
			return false
		}
	}
	return ctx.PHP.AtLeast(phpversion.PHP73) || !cooHasHeredoc(b)
}

// cooHasHeredoc reports whether n contains a heredoc or nowdoc.
func cooHasHeredoc(n syntax.Node) bool {
	found := false
	syntax.Inspect(n, func(n syntax.Node) bool {
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

// cooWordByte reports whether c can be part of an identifier or number.
func cooWordByte(c byte) bool {
	return c == '_' || c >= 0x80 || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// cooOperandText is the operand's text for the other side: an operand
// binding looser than the comparison is parenthesized, or
// `false !== $r = f()` would become `$r = f() !== false`.
func cooOperandText(ctx *analysis.Context, e syntax.Expr) string {
	if astquery.NeedsParensAsEqualityOperand(e) {
		return "(" + ctx.Text(e) + ")"
	}
	return ctx.Text(e)
}
