package astquery

import "custos/internal/php/syntax"

// NullsafeShortCircuits reports whether a null-safe step in the member
// chain of e can make the whole chain evaluate to null without reaching
// its last member (`$a?->b->c()`). Parentheses end the chain: in
// `($a?->b)->c()` the call runs on the null.
func NullsafeShortCircuits(e syntax.Expr) bool {
	for {
		switch x := e.(type) {
		case *syntax.MethodCall:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.PropertyFetch:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.ArrayDimFetch:
			e = x.Var
		default:
			return false
		}
	}
}
