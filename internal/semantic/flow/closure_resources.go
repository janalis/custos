package flow

import "custos/internal/php/syntax"

// Only identities in captures/arguments and a proven freshly returned resource
// cross an invocation boundary. Other callee-local states remain private.
func projectClosureResources(at syntax.Node, result *Value, seeds map[string]Value, caller *frame, owner, body *scopeResult) map[uint32]State {
	projected := map[uint32]State{}
	for _, value := range seeds {
		id := value.Identity
		if id == 0 {
			continue
		}
		state := body.exitStates[id]
		prior := caller.states[id]
		if state != prior && !state.closeChecked && (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") {
			state = State{}
		}
		projected[id] = state
	}
	id := result.Identity
	if id == 0 {
		return projected
	}
	if _, borrowed := projected[id]; borrowed {
		return projected
	}
	result.Identity = 0
	if !body.allocations[id] {
		return projected
	}
	fresh := at.Span().Start + 1
	result.Identity = fresh
	state := body.exitStates[id]
	if !state.closeChecked && (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") {
		state = State{}
	}
	projected[fresh] = state
	if owner.allocations == nil {
		owner.allocations = map[uint32]bool{}
	}
	owner.allocations[fresh] = true
	return projected
}
