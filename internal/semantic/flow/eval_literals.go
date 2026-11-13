package flow

import "custos/internal/php/syntax"

// evalArray evaluates an array literal, its keys and values in order.
func (e *Env) evalArray(n *syntax.Array, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: n}
	v.deep = true
	for _, item := range n.Items {
		if item == nil {
			continue
		}
		e.eval(item.Key, f, r, depth+1)
		value := e.eval(item.Value, f, r, depth+1)
		v = merge(v, value)
		if item.ByRef || item.Unpack && !e.completeSpread(value) {
			v.Complete = false
		}
		if item.Unpack && !value.plain {
			// An object spread into the array is iterated.
			e.implicit(n, f, r)
		}
		if item.ByRef {
			f.alias(rootVariable(item.Value), aliasAny)
		}
	}
	v.Expr = n
	v.elements = e.literalElements(n, r)
	return v
}

// evalInterpolatedString evaluates the parts of an interpolated string.
func (e *Env) evalInterpolatedString(n *syntax.InterpolatedString, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: n}
	// Each part is converted to text before the next one is read.
	for i, part := range n.Parts {
		value := e.eval(part, f, r, depth+1)
		if !value.plain {
			e.implicit(n, f, r)
		}
		if i == 0 {
			v = value
		} else {
			v = merge(v, value)
		}
	}
	v.Expr = n
	return v
}
