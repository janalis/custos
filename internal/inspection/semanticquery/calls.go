package semanticquery

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// ArgBindsByRef reports whether an argument of list binds to a by-reference
// parameter of params (positional, variadic tail or named). With
// callTimeRef, a call-time `&$x` argument counts too.
func ArgBindsByRef(list *syntax.ArgList, params []index.Param, callTimeRef bool) bool {
	if list == nil {
		return false
	}
	pos := 0
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			continue
		}
		if callTimeRef && arg.ByRef {
			return true
		}
		if arg.Name != nil {
			for _, p := range params {
				if p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
					return true
				}
			}
			continue
		}
		// Positional arguments bind in order; one past the last parameter
		// binds to a variadic last parameter, already checked at its own
		// position (positional arguments cannot follow named ones).
		if pos < len(params) && params[pos].ByRef {
			return true
		}
		pos++
	}
	return false
}
