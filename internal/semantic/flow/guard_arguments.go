package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// guardArguments retains written expressions in declaration order. Guard facts
// must apply to the subject parameter even when PHP named arguments are reordered.
// A builtin reference output parameter is legal; explicit reference arguments,
// spreads and malformed bindings cannot establish a validation contract.
func (e *Env) guardArguments(call *syntax.FuncCall) ([]*syntax.Arg, bool) {
	params, _, known := e.declaration(call)
	if !known {
		return nil, false
	}
	args := make([]*syntax.Arg, len(params))
	position := 0
	named := false
	for _, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.ByRef {
			return nil, false
		}
		target := position
		if arg.Name != nil {
			if e.types.PHP < phpversion.PHP80 {
				return nil, false
			}
			named = true
			target = -1
			for i, param := range params {
				if param.Name == arg.Name.Value {
					target = i
					break
				}
			}
		} else {
			if named {
				return nil, false
			}
			position++
		}
		if target < 0 || target >= len(params) || args[target] != nil || params[target].Variadic {
			return nil, false
		}
		args[target] = arg
	}
	for i, param := range params {
		if args[i] == nil && !param.Optional && !param.Variadic {
			return nil, false
		}
	}
	return args, true
}
