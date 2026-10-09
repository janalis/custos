package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// cloneType handles both unary cloning and PHP 8.5's function-like form.
func (e *Env) cloneType(n *syntax.Clone) types.Type {
	if isFirstClassCallable(n.Args) {
		return types.WithCallableReturn(types.Of(`\Closure`), `\Closure`, types.Of("object"))
	}
	return e.TypeOf(n.Expr)
}
