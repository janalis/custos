package flow

import "custos/internal/php/syntax"

// evalClosure binds the captures of a closure. The body executes only on
// invocation, where it can write these names.
func (e *Env) evalClosure(n *syntax.Closure, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: n}
	for _, use := range n.Uses {
		if use.Var != nil {
			e.eval(use.Var, f, r, depth+1)
		}
		if use.ByRef && use.Var != nil {
			f.alias(use.Var.Name, aliasHeld)
			if written := e.holderWrites(n, use.Var.Name); retained(written) {
				if r.holders == nil {
					r.holders = map[string]Value{}
				}
				r.holders[use.Var.Name] = written
			}
		}
	}
	e.captureClosure(n, f, r)
	if _, known := r.closures[n]; known {
		v.targets = oneCallableTarget(n)
	}
	return v
}

// evalArrowFunction captures an arrow function now; the body executes only on
// invocation.
func (e *Env) evalArrowFunction(n *syntax.ArrowFunction, f *frame, r *scopeResult) Value {
	v := Value{Complete: true, Expr: n}
	e.observeArrow(n, f, r)
	e.captureClosure(n, f, r)
	if _, known := r.closures[n]; known {
		v.targets = oneCallableTarget(n)
	}
	return v
}
