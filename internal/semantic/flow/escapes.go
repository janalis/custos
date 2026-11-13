package flow

import "custos/internal/php/syntax"

// invalidateEscaped separates operation histories across calls with unknown effects.
// escapeSet is what one call let escape: the identities reachable from its
// operands, and whether a nested scope may reach any local.
type escapeSet struct {
	identities []uint32
	all        bool
}

func (e *Env) invalidateEscaped(at syntax.Node, values []Value, f *frame, r *scopeResult) {
	identities := map[uint32]bool{}
	all := false
	for _, value := range values {
		collectRecordIdentities(value.elements, identities)
		if value.Identity != 0 {
			identities[value.Identity] = true
		}
		if value.Expr == nil {
			continue
		}
		syntax.Inspect(value.Expr, func(child syntax.Node) bool {
			if !r.complete {
				return false
			}
			r.operations++
			if r.operations > MaxTransfers {
				r.complete = false
				return false
			}
			if syntax.IsVariableScope(child) {
				all = true
				return false
			}
			if expr, ok := child.(syntax.Expr); ok {
				if v := r.values.at(expr); v.Identity != 0 {
					identities[v.Identity] = true
				}
			}
			// An evaluated call already collected what its operands let
			// escape: its subtree holds nothing more.
			if memo, ok := r.escaped[child]; ok {
				all = all || memo.all
				for _, id := range memo.identities {
					identities[id] = true
				}
				return false
			}
			return true
		})
	}
	if r.escaped == nil {
		r.escaped = map[syntax.Node]escapeSet{}
	}
	memo := escapeSet{all: all}
	if len(identities) > 0 {
		memo.identities = make([]uint32, 0, len(identities))
		for id := range identities {
			memo.identities = append(memo.identities, id)
		}
	}
	r.escaped[at] = memo
	if all {
		for _, valueCell := range f.vars {
			value := valueCell.Value
			collectRecordIdentities(value.elements, identities)
		}
	}
	for id := range identities {
		delete(f.states, id)
	}
	for name, valueCell := range f.vars {
		value := valueCell.Value
		if value.Identity != 0 && (all || identities[value.Identity]) {
			value.Invalidated = at.Span().Start + 1
			f.setVar(name, value)
			delete(f.states, value.Identity)
		}
	}
}

// suspend hands control to the consumer of a generator, which may run any
// code before resuming it: like a call of unknown code given what is yielded.
// Locals keep their values, except storage code outside can reach. A
// generator returning by reference also hands out the variable it yields,
// which the consumer can then write whenever the generator is suspended.
func (e *Env) suspend(x syntax.Expr, f *frame, r *scopeResult, depth int) {
	var yielded []Value
	syntax.Children(x, func(child syntax.Node) {
		if expr, ok := child.(syntax.Expr); ok {
			yielded = append(yielded, e.eval(expr, f, r, depth+1))
		}
	})
	if y, ok := x.(*syntax.Yield); ok && y.Value != nil && byReference(syntax.EnclosingVariableScope(x)) {
		f.alias(rootVariable(y.Value), aliasHeld)
	}
	f.unmatchAll(r)
	leaving(f, r)
	e.invalidateEscaped(x, yielded, f, r)
	reenter(x, f, r)
	e.invalidateShared(x, f, r)
	for id := range f.states {
		if id >= bufferID {
			delete(f.states, id)
		}
	}
}

// byReference reports a function-like that returns by reference.
func byReference(scope syntax.Node) bool {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.ByRef
	case *syntax.Method:
		return s.ByRef
	case *syntax.Closure:
		return s.ByRef
	}
	return false
}

// exposed reports whether the frame holds any binding code outside the scope
// can replace. Most frames do not, which spares resolving the callee.
func (e *Env) exposed(f *frame, r *scopeResult) bool {
	return len(f.aliased) > 0 || r.global && e.globalsTable()
}

// invalidateShared discards the bindings code outside the scope can replace
// while a callee with unknown effects, or a callback, runs.
func (e *Env) invalidateShared(at syntax.Node, f *frame, r *scopeResult) {
	// What is left there was possibly written by the code at: a guard read
	// before it says nothing of a read after it.
	stamp := at.Span().Start + 1
	if r.global && e.globalsTable() {
		f.related = nil
		r.operations += len(f.vars)
		for name, oldCell := range f.vars {
			old := oldCell.Value
			v := unsettled(old, e.held(name, r), r)
			v.Invalidated, v.replaced = stamp, true
			f.setVar(name, v)
		}
		return
	}
	// What a holder writes to one name is read under the names bound with it.
	if len(f.aliased) > 0 {
		f.related = nil
	}
	r.operations += len(f.aliased)
	var groups map[uint32]Value
	for name, group := range f.aliased {
		if written := e.held(name, r); group >= aliasGroup && retained(written) {
			if groups == nil {
				groups = map[uint32]Value{}
			}
			groups[group] = effects(groups[group], written, r)
		}
	}
	for name, group := range f.aliased {
		written, shared := groups[group]
		if !shared {
			written = e.held(name, r)
		}
		v := unsettled(f.vars.value(name), written, r)
		v.Invalidated, v.replaced = stamp, true
		f.setVar(name, v)
	}
}

// held is the request data a reference holder of name writes to it: a
// closure of the scope capturing it by reference, or a function of the file
// that binds the same global variable.
func (e *Env) held(name string, r *scopeResult) Value {
	written := r.holders[name]
	if r.global || r.globalNames[name] {
		e.globals()
		if other, ok := e.globalTaint[name]; ok {
			written = effects(written, other, r)
		}
	}
	return written
}

// holderWrites is the request data a scope may leave in name, a variable it
// shares with its creator or the file: what the variable holds when the scope
// ends, and what the scope stores in the globals table entry of that name.
func (e *Env) holderWrites(scope syntax.Node, name string) Value {
	r := e.scope(scope)
	written := Value{Complete: true}
	left, captured := r.exit.lookup(received(name))
	if !captured {
		left = r.exit.value(name)
	}
	for _, v := range []Value{left, r.globalWrites[name]} {
		if retained(v) {
			written = effects(written, v, r)
		}
	}
	written.origins = r.producers(written)
	written.Expr = nil
	return written
}

// release applies the destructor that may run when a variable is given
// another value: what it held may be an object, or hold one, which nothing
// else refers to.
func (e *Env) release(at syntax.Node, name string, f *frame, r *scopeResult) {
	held, tracked := f.vars.lookup(name)
	if tracked && !held.deep || !tracked && !external(name) && !(f.definite && importable(name)) {
		e.implicit(at, f, r)
	}
}
