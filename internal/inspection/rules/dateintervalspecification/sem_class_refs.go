package dateintervalspecification

import (
	"custos/internal/php/syntax"
)

// semArgValue returns the value of a plain argument entry.
func semArgValue(a syntax.Expr) syntax.Expr {
	if arg, ok := a.(*syntax.Arg); ok {
		return arg.Value
	}
	return nil
}
