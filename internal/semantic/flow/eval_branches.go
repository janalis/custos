package flow

import "custos/internal/php/syntax"

// evalTernary evaluates a ternary, the short form included, on both paths.
func (e *Env) evalTernary(n *syntax.Ternary, f *frame, r *scopeResult, depth int) Value {
	var v Value
	cond := e.eval(n.Cond, f, r, depth+1)
	if truth, known := constantTruth(n.Cond); known {
		chosen := n.Else
		if truth {
			chosen = n.Then
		}
		e.refine(n.Cond, f, truth)
		v = cond
		if chosen != nil {
			v = e.eval(chosen, f, r, depth+1)
		}
		return v
	}
	other := clone(f)
	a := cond // The short form yields its condition.
	if n.Then != nil {
		e.refine(n.Cond, f, true)
		a = e.eval(n.Then, f, r, depth+1)
	}
	e.refine(n.Cond, &other, false)
	b := e.eval(n.Else, &other, r, depth+1)
	*f = rejoin(*f, other, r)
	v = merge(a, b)
	return v
}

// evalMatch evaluates a match: the subject, each arm's conditions in order and
// the arm that runs.
func (e *Env) evalMatch(n *syntax.Match, f *frame, r *scopeResult, depth int) Value {
	v := Value{Complete: true, Expr: n}
	// Only an arm result becomes the value; the subject merely selects
	// it. The conditions are compared with it in the order written, each
	// once the ones before it failed, and the default arm, wherever it
	// is written, runs after all of them. Under `match (true)` a
	// condition that matches holds, and one that is a test and does not
	// match failed; `match (false)` reads them the other way round.
	e.eval(n.Cond, f, r, depth+1)
	subject := constantName(n.Cond)
	rest := clone(f)
	var fallback *syntax.MatchArm
	arms := 0
	result := func(arm *syntax.MatchArm, branch frame) {
		body := e.eval(arm.Body, &branch, r, depth+1)
		if e.ends(arm.Body) {
			// An arm that throws or exits yields nothing and goes
			// nowhere after the match.
			return
		}
		if arms++; arms == 1 {
			v, *f = body, branch
		} else {
			v = merge(v, body)
			*f = rejoin(*f, branch, r)
		}
	}
	for _, arm := range n.Arms {
		if len(arm.Conds) == 0 {
			fallback = arm
			continue
		}
		var branch frame
		for i, c := range arm.Conds {
			e.eval(c, &rest, r, depth+1)
			r.copies += len(rest.vars) + len(rest.states)
			matched := clone(&rest)
			if subject == "true" || subject == "false" {
				e.refine(c, &matched, subject == "true")
				if e.boolean(c) {
					e.refine(c, &rest, subject != "true")
				}
			}
			if i == 0 {
				branch = matched
			} else {
				branch = rejoin(branch, matched, r)
			}
		}
		result(arm, branch)
	}
	if fallback != nil {
		result(fallback, rest)
	}
	v.Expr = n
	v.Identity = 0
	return v
}
