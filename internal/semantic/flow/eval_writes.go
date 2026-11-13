package flow

import "custos/internal/php/syntax"

// rebind handles a statement that makes a variable name stand for other
// storage (`unset`, `static`, `global`). Anything but a plain variable is a
// write to what the target names.
func rebind(target syntax.Expr, f *frame, r *scopeResult) {
	if name := plainVariable(target); name != "" {
		f.unmatch(r, name)
		f.detach(name)
		return
	}
	kill(target, f, r)
}

// assign evaluates an assignment on the path that performs it. targeted
// tells that the operands of the target were evaluated before.
func (e *Env) assign(n *syntax.Assign, f *frame, r *scopeResult, depth int, targeted bool) Value {
	var v Value
	target, plain := n.Var.(*syntax.Variable)
	plain = plain && target.NameExpr == nil && !n.ByRef
	if plain && n.Op.Kind != syntax.TEqual && !external(target.Name) {
		// A compound assignment combines the variable with the operand.
		old := e.eval(n.Var, f, r, depth+1)
		operand := e.eval(n.Value, f, r, depth+1)
		if n.Op.Kind == syntax.TConcatEqual && !(old.plain && operand.plain) {
			// Appending converts an object to text.
			e.implicit(n, f, r)
		}
		v = compound(n, old, operand, scalar(r, old.Expr, MaxDepth))
		v.interval = arithmetic(n.Op.Kind, old, operand)
		// `??=` keeps or takes a value, `+=` may unite two arrays.
		both := old.deep && operand.deep
		v.plain, v.deep = n.Op.Kind != syntax.TCoalesceEqual || old.plain && operand.plain, both || n.Op.Kind != syntax.TCoalesceEqual && n.Op.Kind != syntax.TPlusEqual
		f.write(r, target.Name, v)
		return v
	}
	if n.ByRef {
		// A reference to a request table or one of its elements reads what
		// the request sent, whatever a switch matched it with.
		f.unmatch(r, rootVariable(n.Value))
	}
	// PHP evaluates the offsets of an element or property target before
	// the value, and the keys of a destructuring after it.
	_, slots := syntax.UnwrapParens(n.Var).(*syntax.Array)
	if _, listed := syntax.UnwrapParens(n.Var).(*syntax.List); listed {
		slots = true
	}
	if !plain && !slots && !targeted {
		e.operands(n.Var, f, r, depth+1)
	}
	v = e.eval(n.Value, f, r, depth+1)
	if slots {
		e.operands(n.Var, f, r, depth+1)
	}
	if !plain && (slots && (!v.plain || !v.deep && nested(n.Var)) || propertyTarget(n.Var) || n.Op.Kind == syntax.TConcatEqual && !v.plain) {
		// A hook or a magic method writes the property; an object is
		// destructured by a method of its class, and appended as text.
		e.implicit(n, f, r)
	}
	switch {
	case plain && n.Op.Kind == syntax.TEqual:
		// What the variable held is released.
		e.release(n, target.Name, f, r)
		// The variable holds the array with the elements tracked for it.
		bound := v
		if _, held := r.statics[target.Name]; held {
			switch given := syntax.UnwrapParens(n.Value).(type) {
			case *syntax.Literal, *syntax.Array, *syntax.InterpolatedString, *syntax.New, *syntax.Closure, *syntax.ArrowFunction:
				// What a static variable is given here is neither false
				// nor null (see leaving).
				bound.NonFalse, bound.NonNull = true, true
			case *syntax.ConstFetch:
				// A cache emptied with null is not left false.
				bound.NonFalse, bound.NonNull = constantName(given) == "null" || constantName(given) == "true", constantName(given) == "true"
			}
		}
		bound.elements = e.elements(n.Value, f, r)
		f.write(r, target.Name, bound)
		if source := plainVariable(n.Value); source != "" && f.aliased[source] == aliasAny {
			// A copy of an array shares the slots references were taken of.
			f.alias(target.Name, aliasAny)
		}
	case n.ByRef:
		e.reference(n, v, f, r)
		v.Complete = false
	default:
		e.store(n, v, f, r)
		if slots && v.deep {
			taken(n.Var, f)
		}
		v.Complete = false
	}
	return v
}

// taken marks the variables a destructuring fills from an array that holds
// no object: each is given an element of it, or null.
func taken(target syntax.Expr, f *frame) {
	for _, item := range targetItems(target) {
		if item == nil || item.ByRef {
			continue
		}
		if name := plainVariable(item.Value); name != "" {
			slot := f.vars.value(name)
			slot.plain, slot.deep = true, true
			f.setVar(name, slot)
		} else {
			taken(item.Value, f)
		}
	}
}

// operands evaluates what a write target computes before the write: its
// offsets, the object and the name of a property, the keys of a
// destructuring. The variable written is not read.
func (e *Env) operands(target syntax.Expr, f *frame, r *scopeResult, depth int) {
	switch n := syntax.UnwrapParens(target).(type) {
	case *syntax.ArrayDimFetch:
		e.operands(n.Var, f, r, depth+1)
		e.computed(n.Dim, f, r, depth)
	case *syntax.PropertyFetch:
		e.eval(n.Var, f, r, depth+1)
		e.computed(n.Name, f, r, depth)
	case *syntax.StaticPropertyFetch:
		e.computed(n.Class, f, r, depth)
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil {
				e.computed(item.Key, f, r, depth)
				e.operands(item.Value, f, r, depth+1)
			}
		}
	case *syntax.Variable:
		if n.NameExpr != nil {
			e.eval(n, f, r, depth+1)
		}
	default:
		// The array or object written into is the result of an expression
		// (`$o->list()[0] = 1`).
		e.eval(target, f, r, depth+1)
	}
}

// computed evaluates a member, method or class name, or an offset of a write
// target, written as an expression.
func (e *Env) computed(name syntax.Expr, f *frame, r *scopeResult, depth int) {
	switch name.(type) {
	case nil, *syntax.Identifier, *syntax.Name, *syntax.ClassLike, *syntax.Literal, *syntax.ConstFetch:
		// Nothing is computed.
	default:
		e.eval(name, f, r, depth+1)
	}
}

// compound is the value a compound assignment leaves in its variable: the
// concatenation for `.=`, a number for the arithmetic operators, either
// operand's content otherwise. Its expression stays readable only when the
// previous value was known.
func compound(n *syntax.Assign, old, operand Value, scalar bool) Value {
	v := merge(old, operand)
	switch n.Op.Kind {
	case syntax.TPlusEqual:
		// Adding to a value that is not an array yields a number.
		if scalar {
			v.Safe |= everyContext
		}
	case syntax.TMinusEqual, syntax.TMulEqual, syntax.TDivEqual, syntax.TModEqual, syntax.TPowEqual, syntax.TSlEqual, syntax.TSrEqual:
		v.Safe |= everyContext
	}
	v = Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, matched: v.matched && n.Op.Kind == syntax.TConcatEqual}
	if old.Expr != nil && n.Op.Kind != syntax.TCoalesceEqual {
		v.Expr = n
	}
	return v
}

// kill discards the values a write target replaces. An element target
// changes part of its array, which keeps the request data it held; a
// property, static property or slot of either may be a reference to a local
// whose reference was stored there.
