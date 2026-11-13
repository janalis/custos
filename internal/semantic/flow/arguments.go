package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// BoundArguments returns values in declaration parameter order. Unknown
// targets, uncertain packs and duplicate bindings are incomplete.
func (e *Env) BoundArguments(at syntax.Node) ([]Value, bool) {
	r := e.scope(syntax.EnclosingVariableScope(at))
	if i, recorded := r.callIndex[at]; recorded {
		return e.bind(at, r.calls[i].Arguments)
	}
	return nil, false
}

// arguments returns the argument list of a call or a construction, nil when
// none is written.
func arguments(call syntax.Node) *syntax.ArgList {
	switch n := call.(type) {
	case *syntax.FuncCall:
		if n == nil {
			return nil
		}
		return n.Args
	case *syntax.MethodCall:
		return n.Args
	case *syntax.StaticCall:
		return n.Args
	case *syntax.New:
		return n.Args
	}
	return nil
}

// parameters maps each argument of a call to the index of the parameter it
// binds, -1 when there is none. Arguments past the last parameter bind it
// when it is variadic. The result extends positions.
func parameters(list *syntax.ArgList, params []index.Param, positions []int) []int {
	position := 0
	for _, x := range list.Args {
		p := position
		if a, ok := x.(*syntax.Arg); ok && a.Name != nil {
			p = -1
			for j, param := range params {
				if param.Name == a.Name.Value {
					p = j
					break
				}
			}
		} else {
			position++
		}
		if last := len(params) - 1; p > last && last >= 0 && params[last].Variadic {
			p = last
		}
		if p >= len(params) {
			p = -1
		}
		positions = append(positions, p)
	}
	return positions
}

func (e *Env) summaryName(at syntax.Node) string { _, name, _ := e.declaration(at); return name }

func (e *Env) bind(at syntax.Node, values []Value) ([]Value, bool) {
	params, _, known := e.declaration(at)
	if !known {
		return nil, false
	}
	var list *syntax.ArgList
	switch n := at.(type) {
	case *syntax.FuncCall:
		list = n.Args
	case *syntax.MethodCall:
		list = n.Args
	case *syntax.StaticCall:
		list = n.Args
	}
	return e.bindTo(list, params, values)
}

func (e *Env) bindTo(list *syntax.ArgList, params []index.Param, values []Value) ([]Value, bool) {
	args, _, known := e.bindValues(list, params, values, false)
	return args, known
}

func (e *Env) bindValues(list *syntax.ArgList, params []index.Param, values []Value, packs bool) ([]Value, []bool, bool) {
	if len(params) > 64 || list != nil && len(list.Args) > 256 {
		return nil, nil, false
	}
	args := make([]Value, len(params))
	seen := make([]bool, len(params))
	position, count, variadicNext := 0, 0, 0
	var elements *elementSet
	named, unpacked, explicitNamed := false, false, false
	put := func(name string, isNamed bool, value Value) bool {
		count++
		if count > 256 {
			return false
		}
		p := position
		if isNamed {
			if e.types.PHP.Below(phpversion.PHP80) {
				return false
			}
			named, p = true, -1
			for j, param := range params {
				if param.Name == name || param.Variadic {
					p = j
					break
				}
			}
		} else {
			if named {
				return false
			}
			position++
			if last := len(params) - 1; last >= 0 && p >= last && params[last].Variadic {
				p = last
			}
		}
		if p < 0 || p >= len(params) || params[p].ByRef || seen[p] && !params[p].Variadic {
			return false
		}
		if params[p].Variadic {
			// Only the pack's values come from the actual arguments: its
			// numeric and named keys are supplied by the call itself.
			if !seen[p] {
				args[p] = Value{Complete: true}
			}
			args[p] = merge(args[p], value)
			if packs {
				key := name
				if !isNamed {
					key = strconv.Itoa(variadicNext)
					variadicNext++
				}
				elements = elements.with(key, value)
				args[p].elements = elements
				args[p].plain = true
			}
		} else {
			args[p] = value
		}
		seen[p] = true
		return true
	}
	names := map[string]bool{}
	if list != nil {
		for i, x := range list.Args {
			a, ok := x.(*syntax.Arg)
			if !ok || a.ByRef || i >= len(values) {
				return nil, nil, false
			}
			if a.Unpack {
				if explicitNamed || e.types.PHP.Below(phpversion.PHP56) {
					return nil, nil, false
				}
				unpacked = true
				// PHPDoc does not prove the runtime keys or their order.
				t := e.types.Native().TypeOf(a.Value)
				if !t.IsArrayLike() || !t.IsSealedShape() {
					return nil, nil, false
				}
				for _, key := range t.ShapeKeys() {
					v, held := values[i].elements.lookup(key.Name)
					isNamed := !types.IsIntKey(key.Name)
					if key.Optional || !held || !uniqueArgument(key.Name, isNamed, names) || !put(key.Name, isNamed, v) {
						return nil, nil, false
					}
				}
				continue
			}
			name := ""
			if a.Name != nil {
				name = a.Name.Value
				explicitNamed = true
			}
			if unpacked && (name == "" || e.types.PHP.Below(phpversion.PHP81)) || !uniqueArgument(name, a.Name != nil, names) || !put(name, a.Name != nil, values[i]) {
				return nil, nil, false
			}
		}
	}
	for i, p := range params {
		if !seen[i] {
			if !p.Optional && !p.Variadic {
				return nil, nil, false
			}
			args[i] = Value{Complete: true, plain: p.Variadic}
		}
	}
	return args, seen, true
}

// uniqueArgument rejects repeated named keys, including in a variadic pack.
func uniqueArgument(name string, isNamed bool, seen map[string]bool) bool {
	if !isNamed {
		return true
	}
	if seen[name] {
		return false
	}
	seen[name] = true
	return true
}

// localBody finds a unique declaration in this file. Producer syntax is kept
// in this environment only; exported summaries remain independent of the AST.
func (e *Env) localBody(name string) syntax.Node {
	if e.localBodies == nil {
		e.localBodies = map[string]syntax.Node{}
		seen := map[string]bool{}
		syntax.InspectFile(e.file, func(n syntax.Node) bool {
			key := ""
			switch n := n.(type) {
			case *syntax.Function:
				if n.Name != nil {
					key = n.Name.Value
					if ns := e.types.Names.Namespace(n.Span().Start); ns != "" {
						key = ns + `\` + key
					}
				}
			case *syntax.Method:
				if n.Name != nil {
					key = e.types.ClassFQN(syntax.EnclosingClass(n)) + "::" + n.Name.Value
				}
			}
			key = strings.ToLower(key)
			if key != "" {
				if seen[key] {
					e.localBodies[key] = nil
				} else {
					e.localBodies[key] = n
					seen[key] = true
				}
			}
			return true
		})
	}
	body := e.localBodies[name]
	if body != nil && unconditional(body) {
		return body
	}
	return nil
}
