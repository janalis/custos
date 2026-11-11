package astquery

import (
	"custos/internal/php/syntax"
)

func ArgValue(e syntax.Expr) syntax.Expr {
	if a, ok := e.(*syntax.Arg); ok {
		return a.Value
	}
	return nil
}
