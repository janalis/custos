package astquery

import (
	"custos/internal/php/syntax"
)

// NeedsParensAsEqualityOperand reports whether e must be parenthesised when
// it becomes an operand of an equality operator (`===`, `!==`, `==`, `!=`):
// binary operations binding as loosely as or more loosely than equality
// (logical, bitwise, null-coalescing and equality operators — the latter are
// non-associative), ternaries, assignments and the low-precedence keyword
// expressions.
func NeedsParensAsEqualityOperand(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TOr, syntax.TXor, syntax.TAnd, syntax.TCoalesce,
			syntax.TBooleanOr, syntax.TBooleanAnd, syntax.TBar, syntax.TCaret,
			syntax.TAmpersand, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TSpaceship:
			return true
		}
		return false
	case *syntax.Ternary, *syntax.Assign, *syntax.Print, *syntax.Yield,
		*syntax.YieldFrom, *syntax.Include, *syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}
