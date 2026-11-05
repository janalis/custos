package flow

import "custos/internal/php/syntax"

// Reductions preserve runtime order, but carry values and possible sinks only;
// callback operations never establish unconditional caller resource state.
func (e *Env) reducedCallback(at syntax.Node, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok {
		return Value{}, false
	}
	bound, seen, known := e.callbackBindings(at, list, values, false)
	if !known || len(bound) != 3 {
		return Value{}, false
	}
	input, callee := bound[0], bound[1]
	if !input.Complete || input.replaced || input.Invalidated != 0 || input.elements == nil || !input.elements.whole || input.elements.optional || len(input.elements.order) != len(input.elements.items) || !callee.Complete || callee.replaced || callee.Invalidated != 0 {
		return Value{}, false
	}
	carry := bound[2]
	if !seen[2] {
		carry = Value{Complete: true, Expr: scalarExpression(constantScalar{k: 'n'}), plain: true, deep: true}
	}
	prior, had := r.closureSinks[at]
	for _, key := range input.elements.order {
		arguments := []Value{carry, input.elements.items[key]}
		forwarded := &syntax.ArgList{}
		for _, value := range arguments {
			expr := value.Expr
			if expr == nil {
				expr = call
			}
			forwarded.Args = append(forwarded.Args, &syntax.Arg{Value: expr})
		}
		r.copies += len(f.vars) + len(f.states)
		branch := clone(f)
		var valid bool
		if callee.returned != nil {
			carry, _, valid = e.invokeReturned(at, callee.returned, forwarded, arguments, &branch, r)
		} else {
			carry, _, valid = e.invokeCallableSet(at, callee, forwarded, arguments, &branch, r)
		}
		if !valid {
			if had {
				r.closureSinks[at] = prior
			} else {
				delete(r.closureSinks, at)
			}
			return Value{}, false
		}
		carry = eraseRecordResources(carry)
		carry.Identity = 0
	}
	carry.Expr = call
	return carry, true
}
