package infer

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

type argumentBindingState uint8

const (
	bindingMissing argumentBindingState = iota
	bindingBound
	bindingUnknown
)

// bindArgument identifies one nonvariadic parameter on successful calls. A
// spread cannot replace an explicit binding without making PHP reject the call,
// but it prevents proof of omission. Invalid and exhausted bindings stay unknown.
func bindArgument(args *syntax.ArgList, params []index.Param, target int, version phpversion.Version) (syntax.Expr, argumentBindingState) {
	if args == nil || target < 0 || target >= len(params) || len(params) > 64 || len(args.Args) > 256 || params[target].Variadic {
		return nil, bindingUnknown
	}
	if version == 0 {
		version = phpversion.Max
	}
	var result syntax.Expr
	var seen uint64
	position, named, spread := 0, false, false
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg == nil || arg.Value == nil {
			return nil, bindingUnknown
		}
		if arg.Unpack {
			if version < phpversion.PHP56 || named {
				return nil, bindingUnknown
			}
			position, spread = -1, true
			continue
		}
		slot := position
		if arg.Name != nil {
			if version < phpversion.PHP80 || spread && version < phpversion.PHP81 {
				return nil, bindingUnknown
			}
			named, slot = true, -1
			for i, param := range params {
				if strings.TrimPrefix(param.Name, "$") == arg.Name.Value {
					slot = i
					break
				}
			}
			if slot < 0 && params[len(params)-1].Variadic {
				continue // an arbitrary named entry of the variadic tail
			}
		} else {
			if named || position < 0 {
				return nil, bindingUnknown
			}
			position++
			if slot >= len(params) {
				if !params[len(params)-1].Variadic {
					// User functions accept extra positional arguments. These
					// do not replace a declared parameter on a successful call.
					continue
				}
				slot = len(params) - 1
			}
		}
		if slot < 0 {
			return nil, bindingUnknown
		}
		if params[slot].Variadic {
			continue
		}
		bit := uint64(1) << slot
		if seen&bit != 0 {
			return nil, bindingUnknown
		}
		seen |= bit
		if slot == target {
			result = arg.Value
		}
	}
	if result != nil {
		return result, bindingBound
	}
	if spread {
		return nil, bindingUnknown
	}
	return nil, bindingMissing
}
