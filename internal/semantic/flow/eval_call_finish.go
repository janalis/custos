package flow

import "custos/internal/php/syntax"

type evaluatedCall struct {
	prepared                 bool
	receiver                 syntax.Expr
	receiverValue            Value
	followed, closedFollowed uint8
	values                   []Value
	written                  []*syntax.Arg
}

// finishCall applies the same contracts to written and resolved callable calls;
// prepared argument values are never evaluated a second time. It merges the
// arguments into the result, hands the call to the handler of its callee
// kind, then settles states and references.
func (e *Env) finishCall(n syntax.Node, args *syntax.ArgList, input *evaluatedCall, f *frame, r *scopeResult) Value {
	receiver, rv, closedFollowed, values := input.receiver, input.receiverValue, input.closedFollowed, input.values
	f.unmatchAll(r)
	leaving(f, r)
	name, builtin, result, documented := e.callee(n)
	if builtin {
		e.observeIntrospection(n, name, args, f, r)
	}
	recordCall(n, name, receiver, values, r)
	out := callOutcome{value: Value{Complete: true, Expr: n.(syntax.Expr)}}
	derived := mergeArguments(n, name, builtin, args, values, &out.value)
	out.bound, out.boundOK = e.bind(n, values)
	first := Value{}
	if out.boundOK && len(out.bound) > 0 {
		first = out.bound[0]
	} else if call, ok := n.(*syntax.FuncCall); ok && builtin && call.Args != nil {
		if ordered, known := e.guardArguments(call); known && len(ordered) > 0 && ordered[0] != nil {
			first = r.values.at(ordered[0])
		}
	} else if !builtin && len(values) > 0 {
		first = values[0]
	}
	if !e.summarizedCall(n, values, closedFollowed, &out, f, r) && !builtin {
		e.unsummarizedCall(n, values, closedFollowed, &out, f, r)
	}
	if builtin {
		e.builtinCall(n, name, args, values, first, result, &out, f, r)
	} else {
		e.userCall(n, args, values, rv, f, r)
	}
	if builtin && (name == "call_user_func" || name == "call_user_func_array") {
		if returned, states, known := e.invokeCallable(n, values, f, r); known {
			out.value = returned
			out.value.Expr = n.(syntax.Expr)
			out.invoked, out.states = true, states
		}
	}
	e.settleCall(n, name, builtin, args, input, first.Identity, &out, f, r)
	// What PHP itself returns is what it declares. An array one of its
	// functions returns holds what the function was given.
	out.value.setMatched(derived && len(out.value.Sources) == 0)
	kind := e.kinds(result, documented)
	out.value.plain = builtin && kind&(kindKnown|kindObject|kindOpaque) == kindKnown
	out.value.deep = out.value.plain && (kind&kindArray == 0 || e.sealed(n, name, args, r))
	if out.immutablePure || builtin && e.keyed(n, name, args) {
		out.value.plain, out.value.deep = true, true
	}
	if builtin {
		e.callbackResult(n, name, args, values, &out.value, f, r)
	}
	return out.value
}

// mergeArguments merges the arguments of a call into its result and reports
// that a PHP function that edits text makes, of matched labels and literals,
// text of the program.
func mergeArguments(n syntax.Node, name string, builtin bool, args *syntax.ArgList, values []Value, out *Value) bool {
	labels, texts := 0, 0
	for i, a := range values {
		if builtin && searchedOnly(name, args.Args[i], i) {
			continue // what is searched for is removed, never copied
		}
		if builtin && !yieldsElements(name) && !keepsElements(name) && name != "array_merge" && name != "array_replace" && name != "array_map" && name != "array_chunk" {
			// An unaudited transform may expose keys in its result. A later
			// element projection cannot reinterpret those as original values.
			a = projectKeys(a)
		}
		if builtin && yieldsElements(name) {
			a = projectValues(a)
		}
		if a.scalars() && builtin && (yieldsElements(name) || name == "array_chunk" && len(values) == 2) {
			// Only the numbers of the array reach the result.
			a.Safe |= everyContext
		}
		if a.matched() {
			labels++
		} else if a.written() {
			texts++
		}
		*out = merge(*out, a)
	}
	out.Expr = n.(syntax.Expr)
	out.Identity = 0
	// Arrays of numbers put together make one; anything else does not.
	out.setUpload(0)
	out.setScalars(out.scalars() && builtin && (name == "array_merge" || name == "array_replace"))
	return builtin && edits(name) && labels > 0 && labels+texts == len(values)
}

// settleCall records the state a call leaves on the resources it reached and
// discards what it may have changed through references, unpacked arguments
// or a callable it invoked. first is the identity of its first argument.
func (e *Env) settleCall(n syntax.Node, name string, builtin bool, args *syntax.ArgList, input *evaluatedCall, first uint32, out *callOutcome, f *frame, r *scopeResult) {
	if out.value.Identity != 0 && !out.invoked {
		f.setState(out.value.Identity, State{Operation: name, Span: n.Span()})
	}
	id := input.receiverValue.Identity
	if input.receiver == nil {
		id = first
	}
	if id != 0 {
		if builtin {
			f.setState(id, State{Operation: name, Span: n.Span()})
		} else {
			delete(f.states, id)
		}
	}
	for i, a := range input.written {
		if a.ByRef || a.Unpack {
			f.forget(r)
		}
		if i < len(input.values) && !builtin && input.values[i].Identity != 0 {
			delete(f.states, input.values[i].Identity)
		}
	}
	// Supported invocation bindings reject every reference parameter.
	if !out.invoked {
		e.invalidateReferences(n, args, f, r)
	}
	if args != nil && !input.prepared {
		e.accepted(n, args, builtin, input.followed, f)
	}
	if out.invoked {
		for identity, state := range out.states {
			if state.Operation == "" {
				delete(f.states, identity)
			} else {
				f.setState(identity, state)
			}
		}
	}
}

// recordCall records the arguments of a call, merged into those of the
// earlier visit of the same call.
func recordCall(n syntax.Node, name string, receiver syntax.Expr, values []Value, r *scopeResult) {
	if i, ok := r.callIndex[n]; ok {
		for j, v := range values {
			if j < len(r.calls[i].Arguments) {
				r.calls[i].Arguments[j] = merge(r.calls[i].Arguments[j], v)
			}
		}
	} else {
		r.callIndex[n] = len(r.calls)
		r.calls = append(r.calls, Call{Node: n, Name: name, Receiver: receiver, Arguments: values})
	}
}

// callbackResult replaces the result of array_map, array_filter or
// array_reduce with what the callback is known to compute.
func (e *Env) callbackResult(n syntax.Node, name string, args *syntax.ArgList, values []Value, result *Value, f *frame, r *scopeResult) {
	switch name {
	case "array_map":
		if mapped, known := e.mappedCallback(n, args, values, f, r); known {
			*result = mapped
		}
	case "array_filter":
		if filtered, known := e.filteredCallback(n, args, values, f, r); known {
			*result = filtered
		}
	case "array_reduce":
		if reduced, known := e.reducedCallback(n, args, values, f, r); known {
			*result = reduced
		}
	}
}
