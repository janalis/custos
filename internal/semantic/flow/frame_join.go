package flow

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
		if ok && equal(v, w) && (len(v.Sources) > 0 || v.Safe == allContexts && v.upload() == 0 && !v.scalars()) {
			// Joining a value with itself leaves it as it is, once a value
			// without request data has the form a join gives it.
			continue
		}
		if len(v.Sources) > 0 && len(w.Sources) > 0 && &v.Sources[0] != &w.Sources[0] {
			work += (len(v.Sources) + len(w.Sources)) / 2
		}
		m, x := merged(v, w)
		if head {
			x.interval = widenInterval(v, x.interval, m.Expr)
		}
		x.correlate(v, w)
		if !ok && b.definite && !reserved(k) {
			// The other path never assigned the variable: it is null there.
			m.NonFalse, m.plain, m.deep = v.NonFalse, v.plain, v.deep
		}
		x.origins = origins(v, w, f.owner)
		if m.ext = derived(v.ext, x); !equal(m, v) {
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
			m.setOrigins(origins(Value{}, v, f.owner))
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
func (x *valueExt) correlate(a, b Value) {
	if a.tested() == b.tested() {
		x.tested = a.tested()
	}
	if a.unless() == b.unless() {
		x.unless = a.unless()
		x.unlessFalse, x.unlessNull = a.unlessFalse() && b.unlessFalse(), a.unlessNull() && b.unlessNull()
		x.skippedFalse, x.skippedNull = a.skippedFalse() && b.skippedFalse(), a.skippedNull() && b.skippedNull()
	}
}
