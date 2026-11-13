package flow

import (
	"custos/internal/php/syntax"
)

// callOutcome is what finishCall establishes about a call while the handler
// of its callee kind runs: the value the call yields, the arguments bound to
// the parameters, and the resource states an invocation left.
type callOutcome struct {
	value                           Value
	bound                           []Value
	states                          map[uint32]State
	boundOK, invoked, immutablePure bool
}

// summarizedCall applies the summary of a callee that has one, and reports
// whether it had one.
func (e *Env) summarizedCall(n syntax.Node, values []Value, closedFollowed uint8, out *callOutcome, f *frame, r *scopeResult) bool {
	s, ok := e.contract(e.summaryName(n))
	if !ok {
		return false
	}
	r.argumentSensitive = r.argumentSensitive || s.argumentSensitive
	if out.boundOK && len(s.defaults) > 0 {
		out.bound, out.boundOK = e.defaultArguments(n, values, s.defaults)
	}
	out.value = Value{Complete: s.Return.Complete, Expr: n.(syntax.Expr), Safe: s.Return.Safe}
	out.value = substitute(s.Return, out.bound, out.boundOK)
	out.value.Expr = n.(syntax.Expr)
	out.value.Safe |= s.Return.Safe
	forwarded := false
	precise := e.preciseSummaryTarget(n, closedFollowed)
	if out.boundOK && s.Fixed && precise {
		result, states, ok := e.pureReturn(n, s.pure, out.bound, nil, out.value, f, r)
		out.immutablePure = ok && result.plain && result.deep && pureReadOnlyInputs([]Value{result}, nil)
		if !ok {
			result, states, ok = e.forwardedReturn(n, s.forward, out.bound, f, r)
		}
		if ok {
			out.value = result
			out.value.Expr = n.(syntax.Expr)
			out.states = states
			out.invoked = len(states) > 0
			forwarded = true
		}
	}
	if out.value.elements != nil && (!s.Fixed || !precise) {
		out.value.elements = nil
	}
	// An overridable or bodiless target may not be the body that runs:
	// its return keeps the sources it may yield but proves nothing more.
	if !precise {
		out.value.Complete = false
	}
	if out.boundOK {
		out.value.returned = e.bindReturned(n, s.returned, out.bound)
	}
	out.value.scalars = s.Return.scalars && len(out.value.Sources) > 0
	if !forwarded && out.boundOK && out.value.Complete && s.Quiet && precise {
		out.value, out.states = e.instantiateRecord(n, out.value, out.bound, f, r)
		out.invoked = len(out.states) > 0
	}
	if out.boundOK && out.value.Complete {
		if identity, states, known := e.instantiateResource(n, s.resource, out.bound, f, r); known {
			out.value.Identity = identity
			out.invoked, out.states = true, states
		}
	}
	if out.boundOK && len(s.effects) > 0 && precise {
		effects := e.resourceEffects(n, s.effects, out.bound, f, r)
		if len(effects) > 0 {
			if out.states == nil {
				out.states = effects
			} else {
				for id, state := range effects {
					out.states[id] = state
				}
			}
			out.invoked = true
		}
	}
	if s.Generator {
		out.value = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
	}
	return true
}

// unsummarizedCall resolves a project callee without a summary: an
// invocable target, a closed set of returns, or an unresolved dispatch.
func (e *Env) unsummarizedCall(n syntax.Node, values []Value, closedFollowed uint8, out *callOutcome, f *frame, r *scopeResult) {
	if returned, states, known := e.invokeCallable(n, values, f, r); known {
		out.invoked, out.states = true, states
		out.value = returned
		out.value.Expr = n.(syntax.Expr)
	} else if returned, known := e.closedReturn(n, values, closedFollowed); known {
		out.value = returned
		out.value.Expr = n.(syntax.Expr)
	} else {
		r.gap("unresolved-dispatch", n.Span())
		// An unresolved callee may validate or escape its arguments.
		out.value = Value{Expr: out.value.Expr}
	}
}

// builtinCall applies what a PHP function does to its result and to the
// states of the frame.
func (e *Env) builtinCall(n syntax.Node, name string, args *syntax.ArgList, values []Value, first Value, result string, out *callOutcome, f *frame, r *scopeResult) {
	contract := transforms[name]
	if contract.decodes {
		out.value = decodedHTML(out.value)
	}
	if !contract.keepsQuotes {
		out.value.Safe &^= HTMLSingleQuoted | HTMLDoubleQuoted
	}
	if len(out.value.Sources) > 0 && scalarResult(result) {
		out.value.Safe |= everyContext
	}
	switch name {
	case "extract", "parse_str":
		e.importBindings(n, name, args, values, f, r)
	case "array_map":
		e.mappedScalars(values, out)
	}
	if len(values) > 0 && values[0].scalars && keepsElements(name) {
		// The result is that array with fewer elements: the other
		// arguments only choose among them.
		out.value = Value{Complete: out.value.Complete, Expr: out.value.Expr, Sources: values[0].Sources, Safe: values[0].Safe, scalars: true}
	}
	// Invoking a closure object runs project code like calling it.
	if name == "__invoke" || name == "call" || e.callbackParameter(n, args, values) {
		invalidateCallbackRecords(values, f, r)
		reenter(n, f, r)
		if e.exposed(f, r) {
			e.invalidateShared(n, f, r)
		}
	} else if e.handed(n, name, args, values) {
		e.implicit(n, f, r)
	}
	out.value.Safe |= contract.safe
	e.resultSafety(n, name, args, values, out, r)
	frameEffects(n, name, first, f)
}

// userCall invalidates what a call to project code may change.
func (e *Env) userCall(n syntax.Node, args *syntax.ArgList, values []Value, rv Value, f *frame, r *scopeResult) {
	e.invalidateEscaped(n, append(values, rv), f, r)
	if !e.quiet(n, args, values) {
		reenter(n, f, r)
	}
	e.invalidateShared(n, f, r)
	for id := range f.states {
		if id >= bufferID {
			delete(f.states, id)
		}
	}
}
