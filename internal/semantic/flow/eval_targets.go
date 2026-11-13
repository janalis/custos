package flow

import "custos/internal/php/syntax"

func kill(target syntax.Expr, f *frame, r *scopeResult) {
	// A property written, or an element of one, leaves the variable the
	// value it is where PHP refuses to make an object for it.
	written := syntax.UnwrapParens(target)
	for {
		element, ok := written.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		written = syntax.UnwrapParens(element.Var)
	}
	_, member := written.(*syntax.PropertyFetch)
	untest(target, f, r.modern && member)
	switch n := syntax.UnwrapParens(target).(type) {
	case nil:
	case *syntax.Variable:
		if n.NameExpr != nil {
			f.forget(r)
		} else {
			f.write(r, n.Name, Value{})
		}
	case *syntax.ArrayDimFetch:
		switch name := rootVariable(n); name {
		case "":
			kill(n.Var, f, r)
		case "GLOBALS":
			global(n, Value{}, false, f, r)
		default:
			f.write(r, name, unsettled(f.vars.value(name), Value{}, r))
			// The slot may be a reference to a variable (`$a = [&$v]`).
			f.through(r, Value{})
		}
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil {
				kill(item.Value, f, r)
			}
		}
	default:
		f.through(r, Value{})
	}
}

// global applies a write to a globals table entry to the variable it stands
// for: every variable at file level, and the ones bound with `global` in a
// function. definite marks a plain assignment of the whole entry. A dynamic
// key may be any of them.
func global(target *syntax.ArrayDimFetch, v Value, definite bool, f *frame, r *scopeResult) {
	// The request tables are entries of the globals table in every scope.
	f.unmatchAll(r)
	name := assigned(target)
	switch {
	case name == "":
		if r.global {
			f.unsettle(r, v)
		} else {
			r.operations += len(r.globalNames)
			for bound := range r.globalNames {
				f.write(r, bound, unsettled(f.vars.value(bound), v, r))
			}
		}
	case !r.global && !r.globalNames[name]:
	case definite:
		f.write(r, name, v)
	default:
		f.write(r, name, unsettled(f.vars.value(name), v, r))
	}
}

// reference handles `target = &source`.
func (e *Env) reference(n *syntax.Assign, v Value, f *frame, r *scopeResult) {
	untest(n.Value, f, false)
	target, plain := syntax.UnwrapParens(n.Var).(*syntax.Variable)
	source := rootVariable(n.Value)
	if plain && target.NameExpr == nil {
		// What the name held is released; the two names now read one value.
		e.release(n, target.Name, f, r)
		if direct, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable); ok && direct.NameExpr == nil {
			f.bind(target.Name, direct.Name, n.Span().Start)
			f.setVar(target.Name, f.vars.value(direct.Name))
			return
		}
		f.alias(target.Name, aliasAny)
		f.alias(source, aliasAny)
		if dynamic, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable); ok && dynamic.NameExpr != nil {
			// The name stands for a variable chosen at run time.
			f.wild = true
		}
		f.write(r, target.Name, unsettled(Value{}, v, r))
		return
	}
	// The reference is kept in an element, a property or a dynamic name. The
	// source keeps its value; an array given the slot changes.
	f.alias(source, aliasAny)
	switch slot := syntax.UnwrapParens(n.Var).(type) {
	case *syntax.Variable:
		f.forget(r)
	case *syntax.ArrayDimFetch:
		if name := rootVariable(slot); name != "" {
			f.alias(name, aliasAny)
			f.setVar(name, unsettled(f.vars.value(name), v, r))
		}
	}
}

// store handles a write to anything but a plain variable. Only the written
// targets change: an element write alters its array, a destructuring its
// slots, a property write no local binding unless a reference was stored
// there. What is written may carry request data to them.
func (e *Env) store(n *syntax.Assign, v Value, f *frame, r *scopeResult) {
	if v.Identity != 0 {
		// A handle kept elsewhere can be operated on without its variable.
		delete(f.states, v.Identity)
	}
	fetch, element := syntax.UnwrapParens(n.Var).(*syntax.ArrayDimFetch)
	name := rootVariable(n.Var)
	switch {
	case !element || name == "":
		referenceTarget(n.Var, f)
		// What the variables a destructuring names held is released.
		for _, released := range rebound(n.Var) {
			e.release(n, released, f, r)
		}
		kill(n.Var, f, r)
		f.through(r, v)
		if tainted(v) {
			spread(n.Var, v, f, r)
		}
	case name == "GLOBALS":
		if tainted(v) {
			e.globalWrite(n.Var, v, r)
		}
		_, whole := fetch.Var.(*syntax.Variable)
		global(fetch, v, whole && n.Op.Kind == syntax.TEqual, f, r)
	default:
		old, tracked := f.vars.lookup(name)
		// A table of the request, or a variable nothing assigned, is no object.
		held := old.plain || external(name) || !tracked && f.definite && importable(name)
		if !held {
			// An offset of an object is written by a method of its class.
			e.implicit(n, f, r)
		}
		array := unsettled(old, v, r)
		// The variable now holds an array with the value in it: neither
		// what it held nor the value written produced it.
		array.origins = nil
		array.plain = held
		array.deep = held && v.deep && (old.deep || !tracked && (external(name) || f.definite && importable(name)))
		// A literal key directly in the array is tracked on its own; a
		// deeper or compound write adds to what the element held.
		first := fetch
		for {
			inner, nested := first.Var.(*syntax.ArrayDimFetch)
			if !nested {
				break
			}
			first = inner
		}
		if key, known := arrayKey(first.Dim); known && !external(name) && len(old.elements.all()) < MaxSources {
			own := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe}
			if first != fetch || n.Op.Kind != syntax.TEqual {
				previous := old
				if tracked, ok := old.elements.lookup(key); ok {
					previous = tracked
				}
				own = unsettled(previous, v, r)
			}
			r.operations += len(old.elements.all())
			array.elements = old.elements.with(key, own)
		}
		f.write(r, name, array)
		f.through(r, v)
	}
}

// scalar reports an expression whose value is known not to be an array: a
// literal, or a step or compound assignment that started from one.
func scalar(r *scopeResult, x syntax.Expr, depth int) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal, *syntax.IncDec, *syntax.InterpolatedString:
		return true
	case *syntax.Assign:
		switch n.Op.Kind {
		case syntax.TPlusEqual, syntax.TCoalesceEqual:
			return depth > 0 && scalar(r, r.values.at(n.Var).Expr, depth-1)
		}
		return n.Op.Kind != syntax.TEqual
	}
	return false
}

// spread adds the request data of a written value to the bindings a target
// stores it in.
func spread(target syntax.Expr, v Value, f *frame, r *scopeResult) {
	if dynamic, ok := syntax.UnwrapParens(target).(*syntax.Variable); ok && dynamic.NameExpr != nil {
		// Any variable, tracked or not, may be the one written.
		f.unsettle(r, v)
		f.importing(v)
		return
	}
	if name := rootVariable(target); name != "" {
		// The names bound with the target read what it was given.
		f.write(r, name, unsettled(f.vars.value(name), v, r))
		return
	}
	for _, item := range targetItems(target) {
		if item != nil {
			spread(item.Value, v, f, r)
		}
	}
}

// propertyTarget reports a write target that is a property, an element of
// one, or a destructuring with such a slot.
func propertyTarget(target syntax.Expr) bool {
	switch n := syntax.UnwrapParens(target).(type) {
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		return true
	case *syntax.ArrayDimFetch:
		return propertyTarget(n.Var)
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil && propertyTarget(item.Value) {
				return true
			}
		}
	}
	return false
}

// globalWrite records request data stored in a literal globals table entry.
func (e *Env) globalWrite(target syntax.Expr, v Value, r *scopeResult) {
	name := assigned(target)
	if name == "" {
		return
	}
	if r.globalWrites == nil {
		r.globalWrites = map[string]Value{}
	}
	r.globalWrites[name] = effects(r.globalWrites[name], v, r)
}

// assigned names the variable a literal globals table entry stands for.
func assigned(target syntax.Expr) string {
	if t, ok := syntax.UnwrapParens(target).(*syntax.ArrayDimFetch); ok {
		if table, ok := t.Var.(*syntax.Variable); ok && table.Name == "GLOBALS" && table.NameExpr == nil {
			if key, known := literalText(t.Dim); known {
				return key
			}
		}
	}
	return ""
}

// nested reports a destructuring target with a slot that is one itself: the
// element it takes apart may be an object.
func nested(target syntax.Expr) bool {
	for _, item := range targetItems(target) {
		if item != nil && len(targetItems(item.Value)) > 0 {
			return true
		}
	}
	return false
}
