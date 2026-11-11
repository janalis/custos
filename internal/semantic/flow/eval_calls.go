package flow

import (
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
