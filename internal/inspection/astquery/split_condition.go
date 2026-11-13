package astquery

import (
	"slices"

	"custos/internal/php/syntax"
)

// SplitCondition flattens condition c into the operands of its chain of
// one symbolic `&&` or `||` operator and returns that operator (0 and c
// alone when there is none): parentheses and one prefix unary operator
// are stripped first, the left-nested chain is flattened (a parenthesized
// right operand and the keyword operators stay single operands), and each
// operand loses its parentheses.
func SplitCondition(c syntax.Expr) ([]syntax.Expr, syntax.TokenKind) {
	e := syntax.UnwrapParens(c)
	if u, ok := e.(*syntax.Unary); ok {
		e = syntax.UnwrapParens(u.Expr)
	}
	b, ok := e.(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TBooleanAnd && b.Op.Kind != syntax.TBooleanOr {
		return []syntax.Expr{e}, 0
	}
	op := b.Op.Kind
	ops := []syntax.Expr{syntax.UnwrapParens(b.Right)}
	left := syntax.UnwrapParens(b.Left)
	for {
		lb, ok := left.(*syntax.Binary)
		if !ok || lb.Op.Kind != op {
			break
		}
		ops = append(ops, syntax.UnwrapParens(lb.Right))
		left = syntax.UnwrapParens(lb.Left)
	}
	ops = append(ops, left)
	slices.Reverse(ops)
	return ops, op
}
