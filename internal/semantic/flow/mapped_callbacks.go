package flow

import (
	"strconv"

	"custos/internal/php/syntax"
)

// Only a complete finite input can establish individual mapped values. Each
// callback is analyzed independently: its side effects do not establish facts
// about execution order or caller resources.
func (e *Env) mappedCallback(at syntax.Node, args *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok {
		return Value{}, false
	}
	bound, _, known := e.callbackBindings(at, args, values, true)
	if !known || len(bound) != 3 {
		return Value{}, false
	}
	callee := bound[0]
	if !callee.Complete || callee.replaced || callee.Invalidated != 0 {
		return Value{}, false
	}
	inputs := []Value{bound[1]}
	extra := bound[2].elements
	if extra != nil {
		if len(extra.items) > 7 {
			return Value{}, false
		}
		for i := range len(extra.items) {
			input, held := extra.lookup(strconv.Itoa(i))
			if !held {
				return Value{}, false
			}
			inputs = append(inputs, input)
		}
	}
	zip := constantName(callee.Expr) == "null"
	keysByInput := make([][]string, len(inputs))
	count := 0
	for i, input := range inputs {
		if !input.Complete || input.replaced || input.Invalidated != 0 || input.elements == nil {
			return Value{}, false
		}
		// Whole sets, unlike retained subsets after mutation or escape, prove
		// that every entry was captured even inside a sealed argument pack.
		if !input.elements.whole || len(inputs) > 1 && input.elements.optional {
			return Value{}, false
		}
		keys := input.elements.order
		if len(inputs) == 1 {
			keys = sortedElementKeys(input.elements)
		} else if len(keys) != len(input.elements.items) {
			return Value{}, false
		}
		keysByInput[i] = keys
		count = max(count, len(keys))
	}
	if count == 0 {
		return Value{Complete: true, Expr: call, plain: true, deep: true, elements: emptyElements}, true
	}
	prior, had := r.closureSinks[at]
	mapped := &elementSet{items: make(map[string]Value, count), width: count, whole: true}
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
			expr := input.Expr
			if expr == nil {
				expr = call
			}
			forwarded.Args = append(forwarded.Args, &syntax.Arg{Value: expr})
		}
		var returned Value
		var known bool
		if zip {
			if len(arguments) == 1 {
				returned = arguments[0]
			} else {
				entries := &elementSet{items: map[string]Value{}, width: len(arguments), whole: true}
				returned = Value{Complete: true, plain: true, deep: true}
				for i, value := range arguments {
					slot := strconv.Itoa(i)
					entries.put(slot, elementValue(value))
					entries.order = append(entries.order, slot)
					returned = merge(returned, value)
				}
				returned.elements = entries
			}
			known = true
		} else if callee.returned != nil {
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
		mapped.put(key, elementValue(eraseRecordResources(returned)))
		deep := result.deep && returned.deep
		result = merge(result, returned)
		result.deep = deep
	}
	if len(inputs) == 1 {
		mapped.order = inputs[0].elements.order
		mapped.optional = inputs[0].elements.optional
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
	out := Value{storedIdentity: v.storedIdentity, record: v.record, literal: v.literal, Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, elements: v.elements, plain: v.plain, deep: v.deep, scalars: v.scalars}
	switch syntax.UnwrapParens(v.Expr).(type) {
	case *syntax.Literal, *syntax.ConstFetch:
		out.Expr = v.Expr
	}
	return out
}
