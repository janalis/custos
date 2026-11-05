package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// Capture immutable values at creation, not at the later invocation. Repeated
// creations in a loop join their possible captures in the containing result.
func (e *Env) captureClosure(expr syntax.Expr, f *frame, r *scopeResult) {
	var names []string
	switch n := expr.(type) {
	case *syntax.Closure:
		if n.ByRef {
			return
		}
		for _, use := range n.Uses {
			if use.Var == nil || use.ByRef {
				return
			}
			names = append(names, use.Var.Name)
		}
	case *syntax.ArrowFunction:
		if n.ByRef {
			return
		}
		captures := e.arrowNames[n]
		if !captures.complete {
			return
		}
		names = captures.names
	}
	if len(names) > 64 {
		return
	}
	r.operations += len(names) + 1
	if r.operations > MaxTransfers {
		r.complete = false
		r.gap("flow-budget", expr.Span())
		return
	}
	if r.closures == nil {
		r.closures = map[syntax.Expr]map[string]Value{}
	}
	captured, earlier := r.closures[expr]
	if !earlier {
		if !retain(r, len(names)+1) {
			return
		}
		captured = map[string]Value{}
	}
	for _, name := range names {
		value := f.vars[name]
		if value.replaced {
			value.Complete = false
		}
		if earlier {
			value = merge(captured[name], value)
		}
		captured[name] = value
	}
	r.closures[expr] = captured
}

// Only local creation sites retained by an unescaped value can be invoked.
// No closure values or capture frames enter immutable project snapshots.
func (e *Env) invokeClosure(at syntax.Node, expr syntax.Expr, captured map[string]Value, args *syntax.ArgList, values []Value, caller *frame, owner *scopeResult) (Value, map[uint32]State, bool) {
	if e.invocationDepth >= MaxDepth {
		return Value{}, nil, false
	}
	params := syntax.VariableScopeParams(expr)
	if len(params) > 64 {
		return Value{}, nil, false
	}
	lastRequired := -1
	for i, p := range params {
		if p.Default == nil && !p.Variadic {
			lastRequired = i
		}
	}
	declarations := make([]index.Param, len(params))
	for i, p := range params {
		// Reference parameters and malformed variadic placement cannot be
		// projected into an independent invocation frame.
		if p.Var == nil || p.ByRef || p.Variadic && i != len(params)-1 {
			return Value{}, nil, false
		}
		declarations[i] = index.Param{Name: p.Var.Name, Optional: p.Default != nil && i > lastRequired, Variadic: p.Variadic}
	}
	bound, seen, known := e.bindValues(args, declarations, values, true)
	if !known {
		return Value{}, nil, false
	}
	seeds := make(map[string]Value, len(captured)+len(bound))
	for name, value := range captured {
		seeds[name] = value
	}
	for i, p := range params {
		if !seen[i] && p.Default != nil {
			value, known := e.closureDefault(p.Default, owner)
			if !known {
				return Value{}, nil, false
			}
			bound[i] = value
		}
		seeds[p.Var.Name] = bound[i]
	}
	// Invocation analysis has its own cache entry only while executing. The
	// ordinary scope query must never reuse these caller-specific bindings.
	previous, had := e.scopes[expr]
	e.invocationDepth++
	defer func() {
		e.invocationDepth--
		if had {
			e.scopes[expr] = previous
		} else {
			delete(e.scopes, expr)
		}
	}()
	r := e.followSeeded(expr, nil, true, nil, seeds, caller.states)
	for reason, span := range r.gaps {
		owner.gap(reason, span)
	}
	owner.operations += r.operations
	owner.copies += r.copies
	if !r.complete || r.generator || len(r.statics) > 0 || owner.operations > MaxTransfers || owner.copies > MaxCopied {
		return Value{}, nil, false
	}
	result := Value{Complete: true}
	if arrow, ok := expr.(*syntax.ArrowFunction); ok {
		result = r.values[arrow.Expr]
	} else {
		for i, value := range r.returns {
			if i == 0 {
				result = value
			} else {
				result = merge(result, value)
			}
		}
	}
	var sinks []Sink
	for _, called := range r.calls {
		sinks = appendSinks(sinks, e.nativeSinks(called))
		sinks = appendSinks(sinks, e.Sinks(called.Node))
	}
	sinks = appendSinks(sinks, e.commands(syntax.VariableScopeBody(expr), r))
	if owner.closureSinks == nil {
		owner.closureSinks = map[syntax.Node][]Sink{}
	}
	owner.closureSinks[at] = appendSinks(owner.closureSinks[at], sinks)
	projected := projectClosureResources(at, &result, seeds, caller, owner, r)
	return result, projected, true
}
