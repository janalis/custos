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
func (e *Env) invokeClosure(at syntax.Node, values []Value, owner *scopeResult) (Value, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok || e.invocationDepth >= MaxDepth {
		return Value{}, false
	}
	callee := owner.values[call.Name]
	if !callee.Complete || callee.replaced || callee.Invalidated != 0 {
		return Value{}, false
	}
	expr := syntax.UnwrapParens(callee.Expr)
	captured, known := owner.closures[expr]
	if !known {
		return Value{}, false
	}
	params := syntax.VariableScopeParams(expr)
	if len(params) > 64 {
		return Value{}, false
	}
	declarations := make([]index.Param, len(params))
	for i, p := range params {
		// Defaults may instantiate objects; variadics need an element pack,
		// not the scalar join used by ordinary summary substitution.
		if p.Var == nil || p.ByRef || p.Variadic || p.Default != nil {
			return Value{}, false
		}
		declarations[i] = index.Param{Name: p.Var.Name}
	}
	bound, known := e.bindTo(call.Args, declarations, values)
	if !known {
		return Value{}, false
	}
	seeds := make(map[string]Value, len(captured)+len(bound))
	for name, value := range captured {
		seeds[name] = value
	}
	for i, p := range params {
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
	r := e.followSeeded(expr, nil, true, nil, seeds)
	for reason, span := range r.gaps {
		owner.gap(reason, span)
	}
	owner.operations += r.operations
	owner.copies += r.copies
	if !r.complete || r.generator || len(r.statics) > 0 || owner.operations > MaxTransfers || owner.copies > MaxCopied {
		return Value{}, false
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
	// Callee-local identities and operation states are not caller resources.
	// In particular, returning a captured handle must not reopen its state.
	result.Identity = 0
	return result, true
}
