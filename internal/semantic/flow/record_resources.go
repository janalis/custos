package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Record contracts retain alias classes, never callee identities or syntax.
type recordSite struct {
	at    syntax.Node
	token int
}

func sameRecord(a, b *resourceReturn) bool { return a == b || a != nil && b != nil && *a == *b }

func recordShape(value Value, change func(Value) Value) Value {
	set := value.elements
	if set == nil || !set.hasResources {
		return value
	}
	out := cloneElements(set)
	out.hasResources = false
	for _, key := range sortedElementKeys(set) {
		entry := recordShape(set.items[key], change)
		out.put(key, change(entry))
	}
	value.elements = out
	return value
}

func eraseRecordResources(value Value) Value {
	value = recordShape(value, func(entry Value) Value { entry.storedIdentity = 0; entry.record = nil; return entry })
	value.storedIdentity, value.record = 0, nil
	return value
}

func collectRecordIdentities(set *elementSet, ids map[uint32]bool) {
	if set == nil || !set.hasResources {
		return
	}
	for _, value := range set.items {
		if value.storedIdentity != 0 {
			ids[value.storedIdentity] = true
		}
		collectRecordIdentities(value.elements, ids)
	}
}

func resourceElement(at syntax.Node, value Value, f *frame, r *scopeResult) Value {
	out := elementValue(value)
	// Ordinary object identities are never promoted by an array projection.
	state := f.states[value.storedIdentity]
	if value.storedIdentity != 0 && value.Complete && !value.replaced && value.Invalidated == 0 && (r.allocations[value.storedIdentity] || (state.Operation == "fclose" || state.Operation == "fread" || state.Operation == "fwrite") && pendingParameter(at, Value{Complete: true, Identity: value.storedIdentity}, f)) {
		out.Identity = value.storedIdentity
	}
	return out
}

func summarizeRecord(scope syntax.Node, r *scopeResult, sum Summary) Value {
	value := sum.Return
	if value.elements == nil || !value.elements.hasResources {
		return value
	}
	if !sum.Fixed || !sum.Quiet || !value.Complete || value.replaced || value.Invalidated != 0 || byReference(scope) {
		return eraseRecordResources(value)
	}
	borrowed := map[uint32]int{}
	for i, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil && !p.ByRef && !p.Variadic && (p.Type == nil || types.FromNode(p.Type, nil).Has("mixed")) {
			borrowed[p.Var.Span().Start+1] = i
		}
	}
	if !pureRecordPass(syntax.VariableScopeBody(scope)) {
		borrowed = nil
	}
	return recordContracts(value, r, borrowed)
}

func recordContracts(value Value, r *scopeResult, borrowed map[uint32]int) Value {
	if !wholeRecord(value.elements) {
		return eraseRecordResources(value)
	}
	tokens := map[uint32]int{}
	return recordShape(value, func(entry Value) Value {
		id := entry.storedIdentity
		entry.storedIdentity = 0
		entry.record = nil
		if id == 0 || !entry.Complete || entry.Invalidated != 0 || entry.replaced {
			return entry
		}
		if r.allocations[id] {
			token, ok := tokens[id]
			if !ok {
				token = len(tokens) + 1
				tokens[id] = token
			}
			state := r.exitStates[id]
			if !state.closeChecked && (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") {
				state = State{}
			}
			entry.record = &resourceReturn{fresh: true, parameter: -1, token: token, operation: state.Operation, checked: state.closeChecked}
		} else if parameter, ok := borrowed[id]; ok {
			entry.record = &resourceReturn{parameter: parameter}
		}
		return entry
	})
}

// A bounded private identity range sits below the reserved global-state IDs.
// Source identities cannot reach it: source IO is bounded far below 4 GiB.
func (e *Env) recordIdentity(at syntax.Node, token int) uint32 {
	key := recordSite{at: at, token: token}
	if id := e.recordSites[key]; id != 0 {
		return id
	}
	if len(e.recordSites) >= 4096 {
		return 0
	}
	if e.recordSites == nil {
		e.recordSites = map[recordSite]uint32{}
	}
	id := bufferID - 1 - uint32(len(e.recordSites))
	e.recordSites[key] = id
	return id
}

func (e *Env) instantiateRecord(at syntax.Node, value Value, args []Value, f *frame, r *scopeResult) (Value, map[uint32]State) {
	var states map[uint32]State
	value = recordShape(value, func(entry Value) Value {
		contract := entry.record
		entry.record = nil
		if contract == nil {
			return entry
		}
		var id uint32
		var state State
		if contract.fresh {
			id = e.recordIdentity(at, contract.token)
			if id == 0 {
				return entry
			}
			state = State{Operation: contract.operation, Span: at.Span(), Successful: contract.checked, closeChecked: contract.checked}
		} else {
			var known bool
			id, _, known = e.instantiateResource(at, contract, args, f, r)
			if !known {
				return entry
			}
			state = f.states[id]
		}
		if contract.fresh {
			if r.allocations == nil {
				r.allocations = map[uint32]bool{}
			}
			r.allocations[id] = true
		}
		entry.storedIdentity = id
		if states == nil {
			states = map[uint32]State{}
		}
		states[id] = state
		return entry
	})
	return value, states
}

func (e *Env) projectClosureRecord(at syntax.Node, value Value, seeds map[string]Value, caller *frame, owner, body *scopeResult) (Value, map[uint32]State) {
	if value.elements == nil || !value.elements.hasResources {
		return value, nil
	}
	if body.runs {
		return eraseRecordResources(value), nil
	}
	ids := map[uint32]int{}
	var args []Value
	for _, seed := range seeds {
		if seed.Identity != 0 {
			ids[seed.Identity] = len(args)
			args = append(args, seed)
		}
	}
	value, states := e.instantiateRecord(at, recordContracts(value, body, ids), args, caller, owner)
	for id := range ids {
		if _, ok := states[id]; ok {
			state := body.exitStates[id]
			if state != caller.states[id] && !state.closeChecked && (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") {
				state = State{}
			}
			states[id] = state
		}
	}
	return value, states
}

func wholeRecord(set *elementSet) bool {
	if set == nil || !set.hasResources {
		return true
	}
	if !set.whole || set.optional {
		return false
	}
	for _, entry := range set.items {
		if !wholeRecord(entry.elements) {
			return false
		}
	}
	return true
}

func pureRecordPass(body syntax.Node) bool {
	valid := body != nil
	work := 0
	syntax.Inspect(body, func(node syntax.Node) bool {
		work++
		if !valid || work > 256 {
			valid = false
			return false
		}
		switch n := node.(type) {
		case *syntax.Block, *syntax.Return, *syntax.ExprStmt, *syntax.Paren, *syntax.Literal, *syntax.ConstFetch, *syntax.Name:
		case *syntax.Array:
			for _, item := range n.Items {
				if item == nil || item.Unpack || item.ByRef {
					valid = false
				}
			}
		case *syntax.ArrayItem:
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

// A callback may reach records through by-value captures as well as arguments.
// Its per-element analysis does not publish resource effects to the caller.
func invalidateCallbackRecords(values []Value, f *frame, r *scopeResult) {
	ids := map[uint32]bool{}
	for _, value := range values {
		collectRecordIdentities(value.elements, ids)
	}
	for _, value := range f.vars {
		collectRecordIdentities(value.elements, ids)
	}
	for _, captures := range r.closures {
		for _, value := range captures {
			collectRecordIdentities(value.elements, ids)
		}
	}
	for id := range ids {
		delete(f.states, id)
	}
}
