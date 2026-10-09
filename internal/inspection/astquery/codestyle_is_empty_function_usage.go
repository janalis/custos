package astquery

import (
	"custos/internal/php/syntax"
)

// IsEmptyNeedsParens reports whether a comparison replacing n would bind
// looser than n's parent expression.
func IsEmptyNeedsParens(n syntax.Node) bool {
	switch p := n.Parent().(type) {
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor, syntax.TCoalesce:
			return false
		}
		return true
	case *syntax.Unary, *syntax.Instanceof:
		return true
	}
	return false
}
