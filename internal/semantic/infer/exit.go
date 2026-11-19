package infer

import (
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

func (e *Env) exitType(n *syntax.Exit) types.Type {
	ret := types.Of("never")
	if syntax.ExitInvokes(n) {
		return ret
	}
	return types.WithCallableReturn(types.Of(`\Closure`), `\Closure`, ret)
}
