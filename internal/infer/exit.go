package infer

import (
	"custos/internal/syntax"
	"custos/internal/types"
)

func (e *Env) exitType(n *syntax.Exit) types.Type {
	ret := types.Of("never")
	if syntax.ExitInvokes(n) {
		return ret
	}
	return types.WithCallableReturn(types.Of(`\Closure`), `\Closure`, ret)
}
