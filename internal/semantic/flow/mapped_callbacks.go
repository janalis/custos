package flow

import (
	"sort"

	"custos/internal/php/syntax"
)

// Only a complete finite input can establish individual mapped values. Each
// callback is analyzed independently: its side effects do not establish facts
// about execution order or caller resources.
func (e *Env) mappedCallback(at syntax.Node, args *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok || args == nil || len(args.Args) != 2 || len(values) != 2 {
		return Value{}, false
	}
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.ByRef || arg.Unpack {
			return Value{}, false
		}
	}
	ordered, valid := e.types.Native().BuiltinArguments(call)
	bound, known := e.bind(at, values)
	if !valid || !known {
		return Value{}, false
	}
	callee, input := bound[0], bound[1]
	if !callee.Complete || callee.replaced || callee.Invalidated != 0 || !input.Complete || input.replaced || input.Invalidated != 0 || input.elements == nil {
		return Value{}, false
	}
	// Element maps may retain only a subset after a write. Native cardinality
	// must prove that no untracked entries remain before joining all outputs.
	low, high, complete := e.types.Native().TypeOf(ordered[1]).ArrayCardinality()
	if !complete || low != high || low != len(input.elements.items) {
		return Value{}, false
	}
	prior, had := r.closureSinks[at]
	keys := make([]string, 0, len(input.elements.items))
	for key := range input.elements.items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	mapped := &elementSet{items: make(map[string]Value, len(keys))}
	result := Value{Complete: true, Expr: call, plain: true, deep: true}
	for _, key := range keys {
		r.copies += len(f.vars) + len(f.states)
		branch := clone(f)
		value := input.elements.items[key]
		forwarded := &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: ordered[1]}}}
		var returned Value
		var known bool
		if callee.returned != nil {
			returned, _, known = e.invokeReturned(at, callee.returned, forwarded, []Value{value}, r)
		} else {
			returned, _, known = e.invokeCallableSet(at, callee, forwarded, []Value{value}, &branch, r)
		}
		if !known {
			if had {
				r.closureSinks[at] = prior
			} else {
				delete(r.closureSinks, at)
			}
			return Value{}, false
		}
		mapped.items[key] = Value{Complete: returned.Complete, Sources: returned.Sources, Safe: returned.Safe}
		deep := result.deep && returned.deep
		result = merge(result, returned)
		result.deep = deep
	}
	result.Expr, result.Identity, result.elements, result.plain = call, 0, mapped, true
	return result, true
}
