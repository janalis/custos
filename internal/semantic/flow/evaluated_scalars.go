package flow

import (
	"math"

	"custos/internal/php/syntax"
)

// Operands are read from the evaluation snapshot, not reevaluated from their
// syntax. The constant operator evaluator rejects mixed coercions and overflow.
func evaluatedScalar(expr syntax.Expr, value Value, r *scopeResult) Value {
	if !value.Complete || value.replaced || value.Invalidated != 0 {
		return value
	}
	var scalar constantScalar
	var known bool
	switch n := expr.(type) {
	case *syntax.Binary:
		left, lk := knownScalar(r.values.at(n.Left))
		right, rk := knownScalar(r.values.at(n.Right))
		if lk && rk {
			if n.Op.Kind == syntax.TDot && left.k == 's' && right.k == 's' {
				cost := len(left.s) + len(right.s)
				if cost > 4096 || cost > MaxTransfers-r.operations {
					return value
				}
				r.operations += cost
			}
			scalar, known = constantBinary(n.Op.Kind, left, right)
		}
	case *syntax.Unary:
		operand, ok := knownScalar(r.values.at(n.Expr))
		if !ok {
			return value
		}
		switch n.Op.Kind {
		case syntax.TExclaim:
			scalar, known = constantScalar{k: 'b', b: !operand.truth()}, true
		case syntax.TPlus:
			scalar, known = operand, operand.k == 'i' || operand.k == 'f'
		case syntax.TMinus:
			if operand.k == 'i' && operand.i != math.MinInt64 {
				scalar, known = constantScalar{k: 'i', i: -operand.i}, true
			}
			if operand.k == 'f' {
				scalar, known = constantScalar{k: 'f', f: -operand.f}, true
			}
		}
	}
	if known {
		literal := scalar
		value.literal = &literal
	}
	return value
}
