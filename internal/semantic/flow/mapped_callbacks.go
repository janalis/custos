package flow

import (
	"sort"
	"strconv"

	"custos/internal/php/syntax"
)

// Only a complete finite input can establish individual mapped values. Each
// callback is analyzed independently: its side effects do not establish facts
// about execution order or caller resources.
func (e *Env) mappedCallback(at syntax.Node, args *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok || args == nil || len(args.Args) < 2 || len(args.Args) > 9 || len(values) != len(args.Args) {
		return Value{}, false
	}
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.ByRef || arg.Unpack || len(values) > 2 && arg.Name != nil {
			return Value{}, false
		}
	}
	bound, known := e.bind(at, values)
	if !known {
		return Value{}, false
	}
	callee := bound[0]
	if !callee.Complete || callee.replaced || callee.Invalidated != 0 {
		return Value{}, false
	}
	inputs := values[1:]
	if len(values) == 2 {
		inputs = []Value{bound[1]}
	}
	keysByInput := make([][]string, len(inputs))
	count := 0
	for i, input := range inputs {
		if !input.Complete || input.replaced || input.Invalidated != 0 || input.elements == nil {
			return Value{}, false
		}
		// Cardinality and order must describe every entry, not a retained subset.
		low, high, complete := e.types.Native().TypeOf(input.Expr).ArrayCardinality()
		if !complete || low != high || low != len(input.elements.items) {
			return Value{}, false
		}
		keys := input.elements.order
		if len(inputs) == 1 {
			keys = make([]string, 0, len(input.elements.items))
			for key := range input.elements.items {
				keys = append(keys, key)
			}
			sort.Strings(keys)
		} else if len(keys) != len(input.elements.items) {
			return Value{}, false
		}
		keysByInput[i] = keys
		count = max(count, len(keys))
	}
	prior, had := r.closureSinks[at]
	mapped := &elementSet{items: make(map[string]Value, count), width: count}
	result := Value{Complete: true, Expr: call, plain: true, deep: true}
	for offset := range count {
		key := strconv.Itoa(offset)
		if len(inputs) == 1 {
			key = keysByInput[0][offset]
		}
		r.copies += len(f.vars) + len(f.states)
		branch := clone(f)
		forwarded := &syntax.ArgList{}
		arguments := make([]Value, len(inputs))
		for i, input := range inputs {
			// A shorter input supplies null, not the last element or an unknown value.
			value := Value{Complete: true, plain: true, deep: true}
			if offset < len(keysByInput[i]) {
				value = input.elements.items[keysByInput[i][offset]]
			}
			arguments[i] = value
			forwarded.Args = append(forwarded.Args, &syntax.Arg{Value: input.Expr})
		}
		var returned Value
		var known bool
		if callee.returned != nil {
			returned, _, known = e.invokeReturned(at, callee.returned, forwarded, arguments, r)
		} else {
			returned, _, known = e.invokeCallableSet(at, callee, forwarded, arguments, &branch, r)
		}
		if !known {
			if had {
				r.closureSinks[at] = prior
			} else {
				delete(r.closureSinks, at)
			}
			return Value{}, false
		}
		mapped.put(key, elementValue(returned))
		deep := result.deep && returned.deep
		result = merge(result, returned)
		result.deep = deep
	}
	if len(inputs) == 1 {
		mapped.order = inputs[0].elements.order
	} else {
		for i := range count {
			mapped.order = append(mapped.order, strconv.Itoa(i))
		}
	}
	result.Expr, result.Identity, result.elements, result.plain = call, 0, mapped, true
	return result, true
}

// Element projections retain immutable value structure, never resource identity,
// callable frames, alias state or guards from the enclosing evaluation.
func elementValue(v Value) Value {
	out := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, elements: v.elements, plain: v.plain, deep: v.deep, scalars: v.scalars}
	switch syntax.UnwrapParens(v.Expr).(type) {
	case *syntax.Literal, *syntax.ConstFetch:
		out.Expr = v.Expr
	}
	return out
}
