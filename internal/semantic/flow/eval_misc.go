package flow

import "custos/internal/php/syntax"

// evalDynamicCode gives up on the rest of the scope at `eval` or an include,
// whose code is not known.
func evalDynamicCode(x syntax.Expr, f *frame, r *scopeResult) Value {
	f.definite = false
	f.related = nil
	f.resetVars()
	f.states = nil
	if r.complete {
		r.cut = x.Span().Start
	}
	r.complete = false
	r.gap("dynamic-code", x.Span())
	return Value{}
}

// evalOther evaluates the operands of any other expression in order:
// literals, constants, tests, `print`, `exit`, `throw`, `clone` and the like.
func (e *Env) evalOther(x syntax.Expr, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: x}
	first := true
	syntax.Children(x, func(child syntax.Node) {
		if expr, ok := child.(syntax.Expr); ok {
			cv := e.eval(expr, f, r, depth+1)
			if first {
				v = cv
				first = false
			} else {
				v = merge(v, cv)
			}
		}
	})
	v.Expr = x
	switch n := x.(type) {
	case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
		v.Safe |= everyContext
	case *syntax.Literal, *syntax.StringPart, *syntax.ConstFetch, *syntax.MagicConst, *syntax.Name, *syntax.Identifier, *syntax.Throw:
	case *syntax.Print, *syntax.Exit:
		// The operand is converted to text.
		if !first && !v.plain {
			e.implicit(x, f, r)
		}
	case *syntax.ClassConstFetch:
		// `X::class` is the name written: nothing is loaded for it,
		// nor for a constant of a class that is loaded.
		if !className(n) && !r.idle() && !e.loaded(n) {
			e.implicit(x, f, r)
		}
	default:
		// A class loads on first use, and `clone` runs a method.
		e.implicit(x, f, r)
	}
	return v
}
