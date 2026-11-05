package infer

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// BuiltinArguments returns source expressions in the resolved native function's
// parameter order. Missing optional arguments are nil; invalid, variadic or
// uncertain bindings prove nothing. Literal array packs retain PHP key order.
func (e *Env) BuiltinArguments(call *syntax.FuncCall) ([]syntax.Expr, bool) {
	if call == nil {
		return nil, false
	}
	function := e.ResolveFunction(call)
	if function == nil || !function.Builtin || !function.Avail.In(e.PHP) || call.Args == nil || len(function.Params) > maxBoundParams || len(call.Args.Args) > maxWrittenArgs {
		return nil, false
	}
	out := make([]syntax.Expr, len(function.Params))
	position := 0
	named, unpacked, explicitNamed := false, false, false
	put := func(name string, value syntax.Expr) bool {
		if value == nil {
			return false
		}
		slot := position
		if name != "" {
			if e.PHP < phpversion.PHP80 {
				return false
			}
			named, slot = true, -1
			for i, param := range function.Params {
				if strings.TrimPrefix(param.Name, "$") == name {
					slot = i
					break
				}
			}
		} else {
			if named {
				return false
			}
			position++
		}
		if slot < 0 || slot >= len(out) || out[slot] != nil || function.Params[slot].Variadic {
			return false
		}
		out[slot] = value
		return true
	}
	for _, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.ByRef {
			return nil, false
		}
		if arg.Unpack {
			if explicitNamed || e.PHP < phpversion.PHP56 {
				return nil, false
			}
			unpacked = true
			pack, ok := syntax.UnwrapParens(arg.Value).(*syntax.Array)
			if !ok || len(pack.Items) > maxWrittenArgs {
				return nil, false
			}
			keys, values, known := literalArgumentPack(pack, e.PHP)
			if !known {
				return nil, false
			}
			for i, key := range keys {
				name := key
				if n, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(n, 10) == key {
					name = ""
				}
				if !put(name, values[i]) {
					return nil, false
				}
			}
		} else {
			name := ""
			if arg.Name != nil {
				name, explicitNamed = arg.Name.Value, true
			}
			if unpacked && (name == "" || e.PHP < phpversion.PHP81) || !put(name, arg.Value) {
				return nil, false
			}
		}
	}
	for i, param := range function.Params {
		if out[i] == nil && !param.Optional && !param.Variadic {
			return nil, false
		}
	}
	return out, true
}

func literalArgumentPack(array *syntax.Array, version phpversion.Version) ([]string, []syntax.Expr, bool) {
	var keys []string
	var values []syntax.Expr
	positions := map[string]int{}
	next := int64(0)
	for _, item := range array.Items {
		if item == nil || item.Unpack || item.ByRef || item.Value == nil {
			return nil, nil, false
		}
		key := strconv.FormatInt(next, 10)
		if item.Key != nil {
			var known bool
			key, known = syntax.LiteralArrayKey(item.Key)
			if !known {
				return nil, nil, false
			}
		}
		if n, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(n, 10) == key {
			if n >= next || len(keys) == 0 && version >= phpversion.PHP83 {
				if n == 1<<63-1 {
					return nil, nil, false
				}
				next = n + 1
			}
		} else if key == "" {
			return nil, nil, false
		}
		if at, exists := positions[key]; exists {
			values[at] = item.Value
		} else {
			positions[key] = len(keys)
			keys, values = append(keys, key), append(values, item.Value)
		}
	}
	return keys, values, true
}
