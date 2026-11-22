package flow

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
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
	if args.FirstClassCallable() {
		// `f(...)` only makes a closure: nothing is called.
		e.captureCallable(n, rv, f, r)
		value := Value{Complete: true, Expr: n.(syntax.Expr), plain: true, deep: true}
		if _, known := r.callables[value.Expr]; known || r.boundMethods[value.Expr] != nil {
			value.setTargets(oneCallableTarget(value.Expr))
		}
		return value
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
	values, written := []Value{}, []*syntax.Arg{}
	if args != nil {
		values = make([]Value, 0, len(args.Args))
		written = make([]*syntax.Arg, 0, len(args.Args))
		for _, x := range args.Args {
			value := e.eval(x, f, r, depth+1)
			if a, ok := x.(*syntax.Arg); ok && (a.Unpack || mapping) {
				// Capture entries before a later argument can replace them.
				value.setElements(e.elements(a.Value, f, r))
			}
			values = append(values, value)
			if a, ok := x.(*syntax.Arg); ok {
				written = append(written, a)
			}
		}
	}
	return values, written
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
