package flow

import (
	"custos/internal/php/syntax"
)

func (e *Env) transfer(n syntax.Node, f *frame, r *scopeResult) {
	remember(r, n, *f)
	switch n := n.(type) {
	case syntax.Expr:
		v := e.eval(n, f, r, 0)
		if label, ok := n.Parent().(*syntax.Case); ok && label.Cond == n {
			// A label is compared loosely with the subject of its switch.
			subject := label.Parent().(*syntax.Switch).Cond
			if constantName(subject) == "" && (!v.plain && !textless(subject) || !r.values.at(subject).plain && !textless(n)) {
				e.implicit(n, f, r)
			}
		}
	case *syntax.Return:
		v := e.eval(n.Expr, f, r, 0)
		if n.Expr == nil {
			v = nullResult()
		}
		if n.Expr != nil && !v.plain && returnsText(syntax.EnclosingVariableScope(n)) {
			// The declared type converts an object to text.
			r.runs = true
		}
		r.returns = append(r.returns, v)
	case *syntax.ExprStmt:
		e.eval(n.Expr, f, r, 0)
	case *syntax.Foreach:
		// The targets are given an element where the body is entered (see
		// each): the path that leaves the loop computes none of them.
		iterated := e.eval(n.Expr, f, r, 0)
		leaving(f, r)
		if !iterated.plain {
			// An object is iterated by the methods of its class.
			e.implicit(n.Expr, f, r)
		}
	case *syntax.Catch:
		// What the variable held is released, like by an assignment.
		f.unmatchAll(r)
		if n.Var != nil {
			e.release(n, n.Var.Name, f, r)
		}
		kill(n.Var, f, r)
	case *syntax.Echo:
		for _, x := range n.Exprs {
			if v := e.eval(x, f, r, 0); !v.plain {
				// An object is converted to text.
				e.implicit(x, f, r)
			}
		}
		leaving(f, r)
		f.setState(outputID, State{Operation: "echo", Span: n.Span()})
	case *syntax.Global:
		for _, x := range n.Vars {
			// The name leaves what it was bound to for the global variable.
			rebind(x, f, r)
			if name := rootVariable(x); name != "" {
				f.alias(name, aliasHeld)
				if r.globalNames == nil {
					r.globalNames = map[string]bool{}
				}
				r.globalNames[name] = true
				f.setVar(name, unsettled(f.vars.value(name), e.held(name, r), r))
			}
		}
	case *syntax.Unset:
		f.unmatchAll(r)
		leaving(f, r)
		for _, x := range n.Vars {
			e.operands(x, f, r, 0)
			// What is released may have a destructor, and a property may
			// have a magic method that unsets it.
			if name := rootVariable(x); name != "" {
				e.release(n, name, f, r)
			} else {
				e.implicit(n, f, r)
			}
			rebind(x, f, r)
		}
	case *syntax.StaticStmt:
		scope := syntax.EnclosingVariableScope(n)
		for _, x := range n.Vars {
			name := x.Var.Name
			known, seen := r.statics[name]
			if seen && known.state == staticHeld {
				// Reached again, in a loop, the name reads what the function
				// left in the variable.
				r.holds(f, name)
			}
			rebind(x.Var, f, r)
			if !seen {
				if r.statics == nil {
					r.statics = map[string]static{}
				}
				known = e.declared(scope)[name]
				if known.state == staticLoose && !r.distrustAll && !r.distrust[name] {
					known.state = staticHeld
				}
				r.statics[name] = known
			}
			// The first invocation starts from the initial value, any other
			// from what the function left (see declared).
			held := known.state == staticHeld
			f.setVar(name, Value{NonFalse: held && known.starts, entry: held, plain: known.plain, deep: known.deep})
		}
	default:
		syntax.Children(n, func(child syntax.Node) {
			if x, ok := child.(syntax.Expr); ok {
				e.eval(x, f, r, 0)
			}
		})
	}
}

// each starts a turn of a loop on the path that enters its body: the
// targets are evaluated, then given a key and an element of what the loop
// iterates.
func (e *Env) each(n *syntax.Foreach, f *frame, r *scopeResult) {
	iterated := r.values.at(n.Expr)
	if n.ByRef || referenceTarget(n.Value, f) {
		// The loop variable stays bound to an element after the loop.
		f.alias(rootVariable(n.Value), aliasAny)
		f.alias(rootVariable(n.Expr), aliasAny)
	}
	e.operands(n.Key, f, r, 0)
	e.operands(n.Value, f, r, 0)
	if len(targetItems(n.Value)) > 0 && !iterated.deep {
		// An element that is an object is destructured by a method of its
		// class.
		e.implicit(n, f, r)
	}
	kill(n.Key, f, r)
	kill(n.Value, f, r)
	// The keys and elements of request data are request data.
	if len(iterated.Sources) > 0 {
		spread(n.Key, projectKeys(Value{Complete: true, Sources: iterated.Sources, Safe: iterated.Safe}), f, r)
		item := projectValues(iterated)
		if item.scalars {
			item.Safe |= everyContext
		}
		spread(n.Value, item, f, r)
		if name := plainVariable(n.Value); name != "" && iterated.upload == uploadTable {
			// Each element of the table is one uploaded file.
			entry := f.vars.value(name)
			entry.upload = uploadEntry
			f.setVar(name, entry)
		}
	}
	if name := plainVariable(n.Key); name != "" && iterated.plain {
		// A key of an array is a number or text: neither false nor null, and
		// no object. Anything else that is iterated chooses its keys.
		key := f.vars.value(name)
		key.NonFalse, key.NonNull, key.plain, key.deep = true, true, true, true
		f.setVar(name, key)
	}
	if name := plainVariable(n.Value); name != "" && iterated.deep && !n.ByRef {
		// An array that holds no object yields none.
		item := f.vars.value(name)
		item.plain, item.deep = true, true
		f.setVar(name, item)
	}
}

func remember(r *scopeResult, n syntax.Node, f frame) {
	key := syntax.Identity(n)
	if old, ok := r.before[key]; ok {
		for id, s := range old {
			if common, ok := agree(s, f.states[id]); ok {
				old[id] = common
			} else {
				delete(old, id)
			}
		}
	} else {
		if !reserve(r, len(f.states)) {
			return
		}
		var m map[uint32]State
		if len(f.states) > 0 {
			m = make(map[uint32]State, len(f.states))
			for id, s := range f.states {
				m[id] = s
			}
		}
		r.before[key] = m
	}
}

// eval evaluates an expression on the frame: it charges the budget, hands the
// expression to the handler of its family, then normalizes what it yields by
// kind and records it. The epilogue stays inline: passing the value to a
// helper copies it on every evaluation and measurably slows the hot path.
