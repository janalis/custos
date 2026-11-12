package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// evalFuncCall calls a function, named or computed.
func (e *Env) evalFuncCall(n *syntax.FuncCall, f *frame, r *scopeResult, depth int) Value {
	var v Value
	if _, named := n.Name.(*syntax.Name); !named {
		// A callee expression runs first: an invoked closure literal
		// binds its reference captures before the call.
		e.eval(n.Name, f, r, depth+1)
	}
	if !e.options.Assertions && e.types.PHP.AtLeast(phpversion.PHP70) && e.types.BuiltinAssert(n) {
		skipped := clone(f)
		v = e.call(n, n.Args, nil, f, r, depth)
		*f = rejoin(skipped, *f, r)
	} else {
		v = e.call(n, n.Args, nil, f, r, depth)
		if e.options.Assertions {
			if condition := e.types.AssertGuard(n); condition != nil {
				e.refine(condition, f, true)
			}
		}
	}
	if e.types.LegacyAssertCode(n) {
		f.definite = false
		f.related = nil
		f.resetVars()
		f.states = nil
		v = Value{}
		r.complete = false
	}
	return v
}

// evalMethodCall calls a method of an object, possibly null-safe.
func (e *Env) evalMethodCall(n *syntax.MethodCall, f *frame, r *scopeResult, depth int) Value {
	if conditional(n) {
		// The name and the arguments are not evaluated when the
		// receiver is null.
		called := clone(f)
		e.computed(n.Name, &called, r, depth)
		v := e.call(n, n.Args, n.Var, &called, r, depth)
		*f = rejoin(*f, called, r)
		return v
	}
	e.computed(n.Name, f, r, depth)
	return e.call(n, n.Args, n.Var, f, r, depth)
}

// evalStaticCall calls a static method of a class.
func (e *Env) evalStaticCall(n *syntax.StaticCall, f *frame, r *scopeResult, depth int) Value {
	e.computed(n.Class, f, r, depth)
	if conditional(n.Class) {
		// The call is not made when a null-safe link before it is null.
		called := clone(f)
		e.computed(n.Name, &called, r, depth)
		v := e.call(n, n.Args, nil, &called, r, depth)
		*f = rejoin(*f, called, r)
		return v
	}
	e.computed(n.Name, f, r, depth)
	return e.call(n, n.Args, nil, f, r, depth)
}

// evalNew instantiates a class, running its constructor.
func (e *Env) evalNew(n *syntax.New, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: n}
	e.computed(n.Class, f, r, depth)
	values := []Value{}
	if n.Args != nil {
		for _, a := range n.Args.Args {
			values = append(values, e.eval(a, f, r, depth+1))
		}
	}
	f.unmatchAll(r)
	leaving(f, r)
	e.invalidateEscaped(n, values, f, r)
	e.invalidateReferences(n, n.Args, f, r)
	if !e.nativeClass(n) {
		reenter(n, f, r)
		if e.exposed(f, r) {
			e.invalidateShared(n, f, r)
		}
	} else if e.handed(n, "", n.Args, values) {
		e.implicit(n, f, r)
	}
	for id := range f.states {
		if id >= bufferID {
			delete(f.states, id)
		}
	}
	v.Identity = n.Span().Start + 1
	return v
}

// evalArg evaluates an argument given to a call.
func (e *Env) evalArg(n *syntax.Arg, f *frame, r *scopeResult, depth int) Value {
	v := e.eval(n.Value, f, r, depth+1)
	if n.Unpack && !v.plain {
		// An object spread over the parameters is iterated.
		e.implicit(n, f, r)
	}
	if n.ByRef || n.Unpack {
		// The values such an argument binds are not followed.
		v = Value{Expr: v.Expr}
	}
	return v
}

func (e *Env) call(n syntax.Node, args *syntax.ArgList, receiver syntax.Expr, f *frame, r *scopeResult, depth int) Value {
	var rv Value
	if receiver != nil {
		rv = e.eval(receiver, f, r, depth+chainDepth(receiver))
	}
	if args != nil && len(args.Args) == 1 {
		if _, callable := args.Args[0].(*syntax.VariadicPlaceholder); callable {
			// `f(...)` only makes a closure: nothing is called.
			e.captureCallable(n, rv, f, r)
			value := Value{Complete: true, Expr: n.(syntax.Expr), plain: true, deep: true}
			if _, known := r.callables[value.Expr]; known || r.boundMethods[value.Expr] != nil {
				value.targets = oneCallableTarget(value.Expr)
			}
			return value
		}
	}
	followed := followTyped
	closedFollowed := followTyped
	if receiver != nil {
		untest(receiver, f, true)
		followed = e.follows(receiver, f)
		closedFollowed = followed
		if followed == followLoose {
			known := closedBinding(receiver, f)
			if known {
				closedFollowed = followTyped
			}
			if r.closedReceivers == nil {
				r.closedReceivers = map[syntax.Node]bool{}
			}
			if earlier, seen := r.closedReceivers[n]; seen {
				known = known && earlier
			}
			r.closedReceivers[n] = known
		}
		if name := plainVariable(receiver); name != "" && name != "this" {
			// What the variable holds wherever the call is reached.
			if r.receivers == nil {
				r.receivers = map[syntax.Node]uint8{}
			}
			r.receivers[n] = max(r.receivers[n], followed)
		}
	}
	values, written := e.callArguments(n, args, f, r, depth)
	return e.finishCall(n, args, &evaluatedCall{receiver: receiver, receiverValue: rv, followed: followed, closedFollowed: closedFollowed, values: values, written: written}, f, r)
}

// callArguments evaluates the written arguments of a call in order and
// returns their values with the arguments that are not placeholders. A
// spread argument, or any argument of a callback-mapping function, keeps the
// entries it held before a later argument can replace them.
func (e *Env) callArguments(n syntax.Node, args *syntax.ArgList, f *frame, r *scopeResult, depth int) ([]Value, []*syntax.Arg) {
	mapping := false
	if call, ok := n.(*syntax.FuncCall); ok {
		name, builtin := e.callName(call)
		mapping = builtin && (name == "array_map" || name == "array_filter" || name == "array_reduce")
	}
	values := []Value{}
	var argumentStorage [8]*syntax.Arg
	written := argumentStorage[:0]
	if args != nil {
		values = make([]Value, 0, len(args.Args))
		for _, x := range args.Args {
			value := e.eval(x, f, r, depth+1)
			if a, ok := x.(*syntax.Arg); ok && (a.Unpack || mapping) {
				// Capture entries before a later argument can replace them.
				value.elements = e.elements(a.Value, f, r)
			}
			values = append(values, value)
			if a, ok := x.(*syntax.Arg); ok {
				written = append(written, a)
			}
		}
	}
	return values, written
}

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
// kind, then settles states and references. The epilogue stays inline.
func (e *Env) finishCall(n syntax.Node, args *syntax.ArgList, input *evaluatedCall, f *frame, r *scopeResult) Value {
	receiver, rv, followed, closedFollowed, values, written := input.receiver, input.receiverValue, input.followed, input.closedFollowed, input.values, input.written
	f.unmatchAll(r)
	leaving(f, r)
	name, builtin, result, documented := e.callee(n)
	if builtin {
		e.observeIntrospection(n, name, args, f, r)
	}
	recordCall(n, name, receiver, values, r)
	out := callOutcome{value: Value{Complete: true, Expr: n.(syntax.Expr)}}
	// A PHP function that edits text makes, of matched labels and literals,
	// text of the program.
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
		if a.scalars && builtin && (yieldsElements(name) || name == "array_chunk" && len(values) == 2) {
			// Only the numbers of the array reach the result.
			a.Safe |= everyContext
		}
		if a.matched {
			labels++
		} else if a.written() {
			texts++
		}
		out.value = merge(out.value, a)
	}
	derived := builtin && edits(name) && labels > 0 && labels+texts == len(values)
	out.value.Expr = n.(syntax.Expr)
	out.value.Identity = 0
	// Arrays of numbers put together make one; anything else does not.
	out.value.upload, out.value.scalars = 0, out.value.scalars && builtin && (name == "array_merge" || name == "array_replace")
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
	if out.value.Identity != 0 && !out.invoked {
		f.setState(out.value.Identity, State{Operation: name, Span: n.Span()})
	}
	id := rv.Identity
	if receiver == nil {
		id = first.Identity
	}
	if id != 0 {
		if builtin {
			f.setState(id, State{Operation: name, Span: n.Span()})
		} else {
			delete(f.states, id)
		}
	}
	for i, a := range written {
		if a.ByRef || a.Unpack {
			f.forget(r)
		}
		if i < len(values) && !builtin && values[i].Identity != 0 {
			delete(f.states, values[i].Identity)
		}
	}
	// Supported invocation bindings reject every reference parameter.
	if !out.invoked {
		e.invalidateReferences(n, args, f, r)
	}
	if args != nil && !input.prepared {
		e.accepted(n, args, builtin, followed, f)
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
	// What PHP itself returns is what it declares. An array one of its
	// functions returns holds what the function was given.
	out.value.matched = derived && len(out.value.Sources) == 0
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

func (e *Env) callName(n syntax.Node) (string, bool) {
	name, builtin, _, _ := e.callee(n)
	return name, builtin
}

// callee resolves a call's folded name, whether PHP itself declares it and,
// when it does, its declared native return type and its documented one.
func (e *Env) callee(n syntax.Node) (string, bool, string, string) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil && f.Avail.In(e.types.PHP) {
			return strings.ToLower(f.FQN), f.Builtin, f.Return, f.DocReturn
		}
	case *syntax.MethodCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			classes := e.types.TypeOf(n.Var).Classes()
			if len(classes) == 1 {
				if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return, m.DocReturn
				}
			}
			return strings.ToLower(id.Value), false, "", ""
		}
	case *syntax.StaticCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			if class, ok := n.Class.(*syntax.Name); ok {
				if m := e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return, m.DocReturn
				}
			}
			return strings.ToLower(id.Value), false, "", ""
		}
	}
	return "", false, "", ""
}
