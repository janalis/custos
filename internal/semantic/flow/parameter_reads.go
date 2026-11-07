package flow

import "custos/internal/php/syntax"

// ParameterRead reports whether evaluating the scope reads a parameter's entry
// value. A negative answer is proof only when known is true. The CFG accounts
// for joins, loops and exceptional paths; incomplete scopes prove nothing.
func (e *Env) ParameterRead(scope syntax.Node, parameter int) (read, known bool) {
	params := syntax.VariableScopeParams(scope)
	if parameter < 0 || parameter >= len(params) || params[parameter].Var == nil || params[parameter].ByRef {
		return false, false
	}
	r := e.scope(scope)
	if !r.complete {
		return false, false
	}
	if r.parameterReads == nil {
		r.parameterReads = map[string]uint8{}
		for name, state := range r.parameterObserved {
			r.parameterReads[name] = state
		}
		for expr := range r.values.slots {
			value := r.values.at(expr)
			v, ok := expr.(*syntax.Variable)
			if !ok || v.NameExpr != nil {
				continue
			}
			state := r.parameterReads[v.Name]
			if mayHideEntry(value) {
				state |= 1
			}
			for _, source := range value.Sources {
				if source.Kind == "parameter" && source.Parameter >= 0 && source.Parameter < len(params) && params[source.Parameter].Var != nil && params[source.Parameter].Var.Name == v.Name {
					state |= 2
				}
			}
			r.parameterReads[v.Name] = state
		}
	}
	state := r.parameterReads[params[parameter].Var.Name]
	return state&2 != 0, state&2 != 0 || state&1 == 0
}

// mayHideEntry reports that a read value might still be a parameter's entry
// value without listing it among its sources. Sources are may-sources that
// survive unknown writes, joins and opaque calls, so an incomplete value that
// lists every source it has (an unknown call result, for one) cannot carry an
// entry value that was never read. Only a binding untracked code may have
// replaced, or a source list cut at MaxSources, can hide one.
func mayHideEntry(value Value) bool {
	return value.replaced || !value.Complete && len(value.Sources) >= MaxSources
}
