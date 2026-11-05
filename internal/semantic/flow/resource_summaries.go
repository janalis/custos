package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// A resource contract contains no local identity, source position or syntax.
// Fresh identities are instantiated at each caller; pure pass-throughs keep the
// caller's resource identity and operation state.
type resourceReturn struct {
	fresh     bool
	parameter int
	operation string
	checked   bool
}

func (e *Env) summarizeResource(scope syntax.Node, r *scopeResult, sum Summary) *resourceReturn {
	value := sum.Return
	var declared syntax.Expr
	switch n := scope.(type) {
	case *syntax.Function:
		declared = n.ReturnType
	case *syntax.Method:
		declared = n.ReturnType
	}
	returned := types.FromNode(declared, nil)
	if !returned.IsUnknown() && !returned.Has("mixed") {
		return nil
	}
	if !sum.Fixed || !value.Complete || value.Identity == 0 || value.Invalidated != 0 || value.replaced || byReference(scope) {
		return nil
	}
	if r.allocations[value.Identity] {
		state := r.exitStates[value.Identity]
		if !sum.Quiet && !state.closeChecked || (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") && !state.closeChecked {
			state = State{}
		}
		return &resourceReturn{fresh: true, parameter: -1, operation: state.Operation, checked: state.closeChecked}
	}
	if !pureResourcePass(syntax.VariableScopeBody(scope)) {
		return nil
	}
	for i, param := range syntax.VariableScopeParams(scope) {
		if param.Var != nil && !param.ByRef && !param.Variadic && param.Var.Span().Start+1 == value.Identity && (param.Type == nil || types.FromNode(param.Type, nil).Has("mixed")) {
			return &resourceReturn{parameter: i}
		}
	}
	return nil
}

func pureResourcePass(body syntax.Node) bool {
	valid := body != nil
	work := 0
	syntax.Inspect(body, func(node syntax.Node) bool {
		work++
		if !valid || work > 256 {
			valid = false
			return false
		}
		switch n := node.(type) {
		case *syntax.Block, *syntax.Return, *syntax.ExprStmt, *syntax.Paren:
		case *syntax.Assign:
			if n.ByRef || n.Op.Kind != syntax.TEqual {
				valid = false
			}
		case *syntax.Variable:
			valid = n.NameExpr == nil && n.Name != "GLOBALS" && n.Name != "this" && !external(n.Name)
		default:
			valid = false
		}
		return valid
	})
	return valid
}

func (e *Env) instantiateResource(at syntax.Node, contract *resourceReturn, args []Value, f *frame, r *scopeResult) (uint32, map[uint32]State, bool) {
	if contract == nil {
		return 0, nil, false
	}
	if contract.fresh {
		identity := at.Span().Start + 1
		if r.allocations == nil {
			r.allocations = map[uint32]bool{}
		}
		r.allocations[identity] = true
		state := State{Operation: contract.operation, Span: at.Span(), Successful: contract.checked, closeChecked: contract.checked}
		return identity, map[uint32]State{identity: state}, true
	}
	if contract.parameter < 0 || contract.parameter >= len(args) {
		return 0, nil, false
	}
	value := args[contract.parameter]
	if !value.Complete || value.replaced || value.Invalidated != 0 || value.Identity == 0 {
		return 0, nil, false
	}
	id := value.Identity
	state := f.states[id]
	// Scope allocations are the strong resource evidence. A native stream
	// operation can also establish that a previously unknown parameter is one.
	if !r.allocations[id] && state.Operation != "fclose" && state.Operation != "fread" && state.Operation != "fwrite" {
		return 0, nil, false
	}
	return id, map[uint32]State{id: state}, true
}

// Effects describe only the state shared by every normal completion. No callee
// identity or syntax survives publication; each use binds caller identities.
type resourceEffect struct {
	parameter int
	operation string
	checked   bool
}

func summarizeResourceEffects(scope syntax.Node, r *scopeResult, sum Summary) []resourceEffect {
	if !r.complete || !sum.Fixed || !sum.Quiet || sum.Generator || byReference(scope) {
		return nil
	}
	var effects []resourceEffect
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var == nil || p.ByRef || p.Variadic {
			return nil
		}
		state := r.exitStates[p.Var.Span().Start+1]
		switch state.Operation {
		case "fclose", "pclose":
			if !state.closeChecked {
				continue
			}
		case "fread", "fwrite":
		default:
			continue
		}
		effects = append(effects, resourceEffect{parameter: i, operation: state.Operation, checked: state.closeChecked})
	}
	return effects
}

func (e *Env) resourceEffects(at syntax.Node, effects []resourceEffect, args []Value, f *frame, r *scopeResult) map[uint32]State {
	if len(effects) == 0 {
		return nil
	}
	// Distinct parameters can name the same handle at a call site. Their symbolic
	// histories do not describe that ordering, so reject overlapping identities.
	ids := map[uint32]bool{}
	for _, value := range args {
		if value.Identity == 0 {
			continue
		}
		if ids[value.Identity] {
			return nil
		}
		ids[value.Identity] = true
	}
	states := map[uint32]State{}
	for _, effect := range effects {
		id, _, known := e.instantiateResource(at, &resourceReturn{parameter: effect.parameter}, args, f, r)
		if !known {
			return nil
		}
		states[id] = State{Operation: effect.operation, Span: at.Span(), Successful: effect.checked, closeChecked: effect.checked}
	}
	return states
}
