package astquery

import (
	"custos/internal/php/syntax"
)

func PlainVariable(e syntax.Expr) (*syntax.Variable, bool) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" {
		return nil, false
	}
	return v, true
}
