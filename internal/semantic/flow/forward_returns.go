package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Exact parameter forwarding is distinct from a provenance dependency: an
// derived expression can depend on a parameter without preserving its value.
type forwardReturn struct {
	parameter        int
	argument, result types.Type
	weakReturn       bool
}

func returnDeclaration(scope syntax.Node) syntax.Expr {
	switch n := scope.(type) {
	case *syntax.Function:
		return n.ReturnType
	case *syntax.Method:
		return n.ReturnType
	case *syntax.Closure:
		return n.ReturnType
	case *syntax.ArrowFunction:
		return n.ReturnType
	}
	return nil
}

func (e *Env) summarizeForward(scope syntax.Node, r *scopeResult, value Value) *forwardReturn {
	if !r.complete || r.generator || r.runs || byReference(scope) || !value.Complete || value.replaced || value.Invalidated != 0 || value.Identity == 0 {
		return nil
	}
	selected := -1
	params := syntax.VariableScopeParams(scope)
	for i, p := range params {
		if p.Var == nil || p.ByRef || p.Variadic {
			return nil
		}
		if p.Var.Span().Start+1 == value.Identity {
			selected = i
		}
	}
	if selected < 0 || !pureResourcePass(syntax.VariableScopeBody(scope)) {
		return nil
	}
	for i, p := range params {
		if i != selected {
			typ := types.FromNode(p.Type, nil)
			if !typ.IsUnknown() && !typ.Has("mixed") {
				return nil
			}
		}
	}
	return &forwardReturn{parameter: selected, argument: types.FromNode(params[selected].Type, nil), result: types.FromNode(returnDeclaration(scope), nil), weakReturn: !syntax.StrictTypes(e.file)}
}

func (e *Env) forwardedReturn(at syntax.Node, contract *forwardReturn, args []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	if contract == nil || contract.parameter >= len(args) {
		return Value{}, nil, false
	}
	value := args[contract.parameter]
	if !value.Complete || value.replaced || value.Invalidated != 0 {
		return Value{}, nil, false
	}
	if scalar, known := knownScalar(value); known {
		if value.literal == nil {
			literal := scalar
			value.literal = &literal
		}
		value = e.constrainType(contract.argument, value, !syntax.StrictTypes(e.file) || e.internalScalarCallback(at))
		value = e.constrainType(contract.result, value, contract.weakReturn)
		return value, nil, true
	}
	if value.elements == nil || !value.elements.whole || value.elements.optional || !wholeRecord(value.elements) || !recordForwardType(contract.argument) || !recordForwardType(contract.result) {
		return Value{}, nil, false
	}
	// Another argument may be converted to text or destroyed at the boundary.
	for i, arg := range args {
		if i != contract.parameter && (!arg.plain || !arg.deep) {
			return Value{}, nil, false
		}
	}
	ids := map[uint32]bool{}
	collectRecordIdentities(value.elements, ids)
	states := map[uint32]State{}
	for id := range ids {
		if !recordResourceEvidence(at, id, f, r) {
			return Value{}, nil, false
		}
		states[id] = f.states[id]
	}
	return value, states, true
}

func recordForwardType(typ types.Type) bool {
	return typ.IsUnknown() || typ.Has("mixed") || typ.OnlyOf("array", "iterable")
}
