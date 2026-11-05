package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// PHP user functions may ignore surplus positional arguments. Keep native APIs
// strict, and never discard references, packs, opaque values or destructors.
func userArguments(list *syntax.ArgList, params []index.Param, values []Value, allowed bool) (*syntax.ArgList, []Value, bool) {
	if list == nil || len(list.Args) <= len(params) || len(params) > 0 && params[len(params)-1].Variadic {
		return list, values, false
	}
	if !allowed || len(list.Args) != len(values) || len(values) > 256 {
		return list, values, false
	}
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg == nil || arg.Name != nil || arg.Unpack || arg.ByRef {
			return list, values, false
		}
		if i >= len(params) {
			value := values[i]
			if !value.Complete || value.replaced || value.Invalidated != 0 || (!value.plain || !value.deep) && value.literal == nil || value.Identity != 0 || value.elements != nil && value.elements.hasResources {
				return list, values, false
			}
		}
	}
	return &syntax.ArgList{Args: list.Args[:len(params)]}, values[:len(params)], true
}
