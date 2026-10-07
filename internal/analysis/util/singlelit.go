package util

import "custos/internal/syntax"

// SingleStringLiteral returns e when it is a quoted string literal without
// interpolation; otherwise it runs PossibleValues on e and returns the only
// quoted, non-interpolated string literal among the results (other values
// are ignored). nil when there is none or more than one, and when the
// discovery result is unknown (an unstable variable on the path).
func SingleStringLiteral(f *syntax.File, e syntax.Expr) syntax.Expr {
	if _, _, ok := QuotedStringRaw(e); ok {
		return e
	}
	if _, ok := e.(*syntax.Literal); ok {
		return nil
	}
	var found syntax.Expr
	for _, v := range PossibleValues(f, e) {
		if _, _, ok := QuotedStringRaw(v); ok {
			if found != nil {
				return nil
			}
			found = v
		}
	}
	return found
}
