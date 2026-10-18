package infer

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// builtinBindingsKnown proves a complete written binding before refining a
// builtin return type. Unknown spreads and invalid calls retain the declaration
// type. A fixed bitset avoids allocating on this hot path.
func builtinBindingsKnown(list *syntax.ArgList, params []index.Param, php phpversion.Version, allowTailSpread bool) bool {
	if list == nil || len(list.Args) > 256 || len(params) > 64 {
		return false
	}
	var seen uint64
	position, named := 0, false
	for slot, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Value == nil || arg.ByRef {
			return false
		}
		// These array contracts retain proven prefix operands when only the
		// optional/variadic tail is unpacked. The callback remains unknown.
		if arg.Unpack {
			if !allowTailSpread || named || php < phpversion.PHP56 || slot != len(list.Args)-1 {
				return false
			}
			break
		}
		target := position
		if arg.Name != nil {
			if php < phpversion.PHP80 {
				return false
			}
			named, target = true, -1
			for i, param := range params {
				if strings.TrimPrefix(param.Name, "$") == arg.Name.Value {
					target = i
					break
				}
			}
		} else {
			if named {
				return false
			}
			position++
			if target >= len(params) && len(params) > 0 && params[len(params)-1].Variadic {
				target = len(params) - 1
			}
		}
		if target < 0 || target >= len(params) {
			return false
		}
		bit := uint64(1) << target
		if seen&bit != 0 && (arg.Name != nil || !params[target].Variadic) {
			return false
		}
		seen |= bit
	}
	for i, param := range params {
		if seen&(uint64(1)<<i) == 0 && !param.Optional && !param.Variadic {
			return false
		}
	}
	return true
}
