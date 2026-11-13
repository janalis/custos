package flow

import (
	"maps"
)

type frame struct {
	owner  *scopeResult
	vars   bindings
	states map[uint32]State
	// aliased names are bound to storage that code outside the tracked
	// bindings can write: aliasHeld for storage only a callee can reach,
	// aliasAny for a reference kept where the analysis does not follow it (a
	// property, an array slot, a loop variable), and a larger number for the
	// names one `$a = &$b` binds together, which a write through either
	// replaces for both. nil for the common scope without such bindings.
	aliased map[string]uint32
	// definite tells that a name without a binding is a variable nothing
	// assigned: it is null, not false. It holds in a function that captures
	// nothing, until a write the analysis cannot name (`$$n = …`, `extract`,
	// an included file) may have created any variable.
	definite bool
	// wild tells that some name is bound to a variable the analysis cannot
	// name (`$r = &$$n`): a write through a reference kept somewhere may
	// reach any variable.
	wild bool
	// cellsShared stays set after detaching a map: unchanged cells may still
	// belong to another frame. A replacement tagged with the current map's
	// sharing marker is private and can be updated without another allocation.
	cellsShared bool
	// related lists what holds between the left operand of a short-circuit
	// operation and the variable its right operand assigns in place (see
	// relation). The list is never modified once a frame holds it.
	related []relation
	// varsShared is the flag of vars, which every frame holding that map
	// holds, however it was copied: set once clone gave the map to two
	// frames that change apart. A frame copies a shared map before its
	// first change. Nil for a map no clone shared yet.
	varsShared *bool
}

// setVar updates an owned cell or replaces one a snapshot may still hold.
func (f *frame) setVar(name string, v Value) {
	if f.cellsShared || f.varsShared != nil && *f.varsShared {
		if held, known := f.vars.lookup(name); known && equal(held, v) {
			return
		}
	}
	f.ownVars()
	if cell := f.vars[name]; cell != nil && (!f.cellsShared || cell.owner != nil && cell.owner == f.varsShared) {
		cell.Value = v
		return
	}
	cell := &bindingCell{Value: v, owner: f.varsShared}
	f.vars[name] = cell
}

// unsetVar removes the binding of name.
func (f *frame) unsetVar(name string) {
	if _, ok := f.vars[name]; ok {
		f.ownVars()
		delete(f.vars, name)
	}
}

func (f *frame) ownVars() {
	if f.varsShared != nil && *f.varsShared {
		f.vars, f.varsShared = maps.Clone(f.vars), f.shareFlag()
		f.cellsShared = true
	}
}

// resetVars gives the frame new, empty variables.
func (f *frame) resetVars() {
	f.vars, f.varsShared = bindings{}, f.shareFlag()
	f.cellsShared = false
}

// clone returns a frame holding what f holds, which either can then change
// apart: their variables are copied on the first change of either.
func clone(f *frame) frame {
	if f.varsShared == nil {
		f.varsShared = f.shareFlag()
	}
	*f.varsShared = true
	f.cellsShared = true
	n := frame{cellsShared: true, owner: f.owner, vars: f.vars, varsShared: f.varsShared, states: maps.Clone(f.states), definite: f.definite, wild: f.wild, related: f.related}
	if len(f.aliased) > 0 {
		n.aliased = make(map[string]uint32, len(f.aliased))
		for k, group := range f.aliased {
			n.aliased[k] = group
		}
	}
	return n
}

// setState records the state of the resource or output identity id; a frame
// without any keeps no map.
func (f *frame) setState(id uint32, s State) {
	if f.states == nil {
		f.states = map[uint32]State{}
	}
	f.states[id] = s
}

// alias records a name whose storage is shared outside the tracked bindings.
// Two different bindings of one name leave only that it is shared somewhere,
// which then holds for every name bound together with it.
func (f *frame) alias(name string, group uint32) {
	if name == "" {
		return
	}
	if f.aliased == nil {
		f.aliased = map[string]uint32{}
	}
	old, known := f.aliased[name]
	switch {
	case known && (group == aliasHeld || group == old):
		return
	case !known || old == aliasHeld:
		f.aliased[name] = group
	default:
		for other, g := range f.aliased {
			if g >= aliasGroup && (g == old || g == group) {
				f.aliased[other] = aliasAny
			}
		}
		f.aliased[name] = aliasAny
	}
	// What related the name to another one may no longer hold.
	f.related = nil
}

// bind makes target a reference to source, as `$target = &$source` does: the
// two names, and the names already bound with source, now share one value.
func (f *frame) bind(target, source string, site uint32) {
	group, known := f.aliased[source]
	if !known || group == aliasHeld {
		group = site + aliasGroup
	}
	if f.aliased == nil {
		f.aliased = map[string]uint32{}
	}
	f.related = nil
	f.aliased[source] = group
	f.aliased[target] = group
}

// write binds a plain variable. A write through a name bound by `=&` gives
// the same value to the names bound with it. A write through a name whose
// reference is kept somewhere the analysis does not follow may reach every
// other such name: what they held is gone, and the request data written may
// now be read under them. The walk is charged to the scope's budget.
func (f *frame) write(r *scopeResult, name string, v Value) {
	// What is written is the binding's own value again, whatever replaced
	// the operands it was computed from.
	v.replaced = false
	f.unmatch(r, name)
	if len(f.related) > 0 {
		f.rewritten(r, name, v)
	}
	if group, shared := f.aliased[name]; shared && group == aliasAny && f.wild {
		f.unsettle(r, v)
	}
	if group, shared := f.aliased[name]; shared && group != aliasHeld {
		r.operations += len(f.aliased)
		for other, g := range f.aliased {
			switch {
			case other == name || g != group:
			case group == aliasAny:
				f.setVar(other, unsettled(f.vars.value(other), v, f.owner))
			default:
				f.setVar(other, v)
			}
		}
	}
	f.setVar(name, v)
}

// detach makes name a variable of its own again, as `unset`, `static` and
// `global` do: the storage it was bound to keeps its value for the other
// names, and the name starts over with nothing known.
func (f *frame) detach(name string) {
	f.related = nil
	delete(f.aliased, name)
	f.setVar(name, Value{})
}

// through is a write to storage the analysis does not follow (a property, an
// array slot): it may reach any variable whose reference is kept there.
func (f *frame) through(r *scopeResult, v Value) {
	if len(f.aliased) == 0 {
		return
	}
	f.related = nil
	r.operations += len(f.aliased)
	for other, g := range f.aliased {
		if g == aliasAny {
			f.setVar(other, unsettled(f.vars.value(other), v, f.owner))
		}
	}
}

// forget discards what is known about every binding and operation, as after
// a write whose target the analysis does not follow. Request data stays.
func (f *frame) forget(r *scopeResult) {
	f.definite = false
	f.unsettle(r, Value{})
	f.states = nil
}

// unsettle makes every binding unknown, keeping its request data and adding
// what was written.
func (f *frame) unsettle(r *scopeResult, written Value) {
	f.related = nil
	r.operations += len(f.vars)
	for name, oldCell := range f.vars {
		old := oldCell.Value
		f.setVar(name, unsettled(old, written, f.owner))
	}
}

// unsettled is the value of a binding that code the analysis does not follow
// may have replaced. Nothing is known about it any more, except that request
// data it held, or that such code writes (written), may still be there:
// unknown must not read as untainted. The same goes for what a parameter
// brought in, which is request data when a caller passes some.
func unsettled(old, written Value, owners ...*scopeResult) Value {
	v := Value{}
	if len(old.Sources) > 0 {
		v = Value{Complete: true, Sources: old.Sources, Safe: old.Safe}
	}
	if len(written.Sources) > 0 {
		if !v.Complete {
			v = Value{Complete: true}
		}
		v = merge(v, Value{Complete: true, Sources: written.Sources, Safe: written.Safe})
	}
	v.origins = origins(Value{}, written, owners...)
	return v
}

// tainted reports a retained request source.
func tainted(v Value) bool {
	for _, s := range v.Sources {
		if s.Parameter < 0 {
			return true
		}
	}
	return false
}

// agree keeps an operation every path performed; its success must hold on each.
func agree(a, b State) (State, bool) {
	if a.Operation != b.Operation || a.Span != b.Span {
		return State{}, false
	}
	a.Successful = a.Successful && b.Successful
	a.closeChecked = a.closeChecked && b.closeChecked
	return a, true
}

// rejoin joins a path that split off inside an expression into the first
// one, charging the frame copied for it and the join to the scope's budget.
func rejoin(first, second frame, r *scopeResult) frame {
	_, work := first.absorb(second)
	r.copies += work + len(second.vars) + len(second.states)
	return first
}

// absorb joins another path into the frame in place. It reports whether the
// frame changed and the work done, in bindings visited and sources compared.
func (f *frame) absorb(b frame) (bool, int) {
	return f.absorbAt(b, false)
}

func (f *frame) absorbAt(b frame, head bool) (bool, int) {
	changed, work, shared := false, len(f.vars)+len(f.states), 0
	// What relates two variables is joined from what each path holds of
	// them, before their values are.
	if len(f.related)+len(b.related) > 0 && f.correlated(b) {
		changed = true
	}
	for k, vCell := range f.vars {
		v := vCell.Value
		w, ok := b.vars.lookup(k)
		if ok {
			shared++
		}
		if ok && equal(v, w) && (len(v.Sources) > 0 || v.Safe == allContexts && v.upload == 0 && !v.scalars) {
			// Joining a value with itself leaves it as it is, once a value
			// without request data has the form a join gives it.
			continue
		}
		if len(v.Sources) > 0 && len(w.Sources) > 0 && &v.Sources[0] != &w.Sources[0] {
			work += (len(v.Sources) + len(w.Sources)) / 2
		}
		m := merge(v, w)
		if head {
			m.interval = widenInterval(v, m)
		}
		m.correlate(v, w)
		if !ok && b.definite && !reserved(k) {
			// The other path never assigned the variable: it is null there.
			m.NonFalse, m.plain, m.deep = v.NonFalse, v.plain, v.deep
		}
		if m.origins = origins(v, w, f.owner); !equal(m, v) {
			f.setVar(k, m)
			changed = true
		}
	}
	if shared < len(b.vars) {
		work += len(b.vars)
		for k, vCell := range b.vars {
			v := vCell.Value
			if _, ok := f.vars[k]; ok {
				continue
			}
			m := merge(Value{}, v)
			m.origins = origins(Value{}, v, f.owner)
			if f.definite && !reserved(k) {
				m.NonFalse, m.plain, m.deep = v.NonFalse, v.plain, v.deep
			}
			f.setVar(k, m)
			changed = true
		}
	}
	for k, v := range f.states {
		if common, ok := agree(v, b.states[k]); !ok {
			delete(f.states, k)
			changed = true
		} else if common != v {
			f.states[k] = common
			changed = true
		}
	}
	if f.definite && !b.definite || !f.wild && b.wild {
		f.definite, f.wild = f.definite && b.definite, true && (f.wild || b.wild)
		changed = true
	}
	// A definite binding one path lacks, or holds otherwise, may not hold.
	for k, group := range f.aliased {
		if other, both := b.aliased[k]; group >= aliasGroup && (!both || other == aliasHeld) {
			f.alias(k, aliasAny)
			changed = true
		}
	}
	for k, group := range b.aliased {
		mine, both := f.aliased[k]
		if group >= aliasGroup && (!both || mine == aliasHeld) {
			// The names it is bound with on the other path may not be either.
			f.alias(k, group)
			group = aliasAny
		}
		f.alias(k, group)
		changed = changed || !both || f.aliased[k] != mine
	}
	return changed, work
}

// correlate gives the join of two paths what both know of a binding through
// a short-circuit operation: that its left operand read the binding last,
// and what holds of the binding unless its right operand was skipped.
func (v *Value) correlate(a, b Value) {
	if a.tested == b.tested {
		v.tested = a.tested
	}
	if a.unless == b.unless {
		v.unless, v.unlessFalse, v.unlessNull = a.unless, a.unlessFalse && b.unlessFalse, a.unlessNull && b.unlessNull
		v.skippedFalse, v.skippedNull = a.skippedFalse && b.skippedFalse, a.skippedNull && b.skippedNull
	}
}

// reserve bounds the operation states remembered per node.
func reserve(r *scopeResult, cells int) bool {
	if cells > MaxTransfers-r.facts {
		r.spent = true
		r.complete = false
		return false
	}
	r.facts += cells
	return true
}

// retain bounds the frame cells kept as block inputs.
func retain(r *scopeResult, cells int) bool {
	if cells > MaxRetained-r.originReserve-r.retained {
		r.spent = true
		r.complete = false
		return false
	}
	r.retained += cells
	return true
}

// retained reports effects useful to either request-data or failure analysis.
func retained(v Value) bool { return tainted(v) || v.Expr != nil || v.origins != nil }

// effects joins local writes, retaining their syntax only in the environment.
func effects(a, b Value, owners ...*scopeResult) Value {
	v := merge(a, b)
	v.origins = origins(a, b, owners...)
	return v
}
