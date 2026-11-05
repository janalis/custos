package flow

import "custos/internal/php/syntax"

const maxCallableTargets = 8

type callableSet struct{ items []syntax.Expr }

func oneCallableTarget(expr syntax.Expr) *callableSet {
	return &callableSet{items: []syntax.Expr{expr}}
}

func sameCallableTargets(a, b *callableSet) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || len(a.items) != len(b.items) {
		return false
	}
	for _, item := range b.items {
		if !hasCallable(a, item) {
			return false
		}
	}
	return true
}

func hasCallable(set *callableSet, expr syntax.Expr) bool {
	for _, item := range set.items {
		if item == expr {
			return true
		}
	}
	return false
}

func joinCallableTargets(a, b *callableSet) *callableSet {
	if a == nil || b == nil {
		return nil
	}
	extra := 0
	for _, item := range b.items {
		if !hasCallable(a, item) {
			extra++
		}
	}
	if extra == 0 {
		return a
	}
	if len(a.items)+extra > maxCallableTargets {
		return nil
	}
	result := &callableSet{items: append([]syntax.Expr(nil), a.items...)}
	for _, item := range b.items {
		if !hasCallable(result, item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func applyCallStates(f *frame, states map[uint32]State) {
	for id, state := range states {
		if state.Operation == "" {
			delete(f.states, id)
		} else {
			f.setState(id, state)
		}
	}
}

func (e *Env) invokeCallableSet(at syntax.Node, callee Value, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	targets := []syntax.Expr{callee.Expr}
	if callee.targets != nil {
		targets = callee.targets.items
	}
	prior, had := r.closureSinks[at]
	var result Value
	var joined frame
	changed := map[uint32]bool{}
	var sinks []Sink
	for i, target := range targets {
		r.copies += len(f.vars) + len(f.states)
		r.operations++
		if r.copies > MaxCopied || r.operations > MaxTransfers {
			r.complete = false
			r.gap("flow-budget", at.Span())
			if had {
				r.closureSinks[at] = prior
			} else {
				delete(r.closureSinks, at)
			}
			return Value{}, nil, false
		}
		branch := clone(f)
		if r.closureSinks != nil {
			delete(r.closureSinks, at)
		}
		value, states, known := e.invokeCallableTarget(at, target, list, values, &branch, r)
		if !known {
			if had {
				r.closureSinks[at] = prior
			} else {
				delete(r.closureSinks, at)
			}
			return Value{}, nil, false
		}
		sinks = appendSinks(sinks, r.closureSinks[at])
		applyCallStates(&branch, states)
		for id := range states {
			changed[id] = true
		}
		if i == 0 {
			result, joined = value, branch
		} else {
			result = merge(result, value)
			joined = rejoin(joined, branch, r)
		}
	}
	*f = joined
	projected := map[uint32]State{}
	for id := range changed {
		projected[id] = joined.states[id]
	}
	// Every supported target publishes a sink list, including an empty one.
	r.closureSinks[at] = appendSinks(prior, sinks)
	return result, projected, true
}

// Literal names can participate in a finite target set before resolution.
// Every member must still resolve and bind at the eventual invocation.
func joinCallableValues(a, b Value) *callableSet {
	left, right := a.targets, b.targets
	if left == nil {
		if literal, ok := syntax.UnwrapParens(a.Expr).(*syntax.Literal); ok && literal.LitKind == syntax.LitString {
			left = oneCallableTarget(literal)
		}
	}
	if right == nil {
		if literal, ok := syntax.UnwrapParens(b.Expr).(*syntax.Literal); ok && literal.LitKind == syntax.LitString {
			right = oneCallableTarget(literal)
		}
	}
	return joinCallableTargets(left, right)
}
