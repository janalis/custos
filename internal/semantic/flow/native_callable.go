package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// Use a source-positioned view of the actual invocation with a resolved native
// name. Its argument nodes still belong to the original tree and are read only.
func (e *Env) invokeNativeCallable(at syntax.Node, args *syntax.ArgList, values []Value, name string, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok || e.invocationDepth >= MaxDepth {
		return Value{}, nil, false
	}
	r.operations++
	if r.operations > MaxTransfers {
		r.complete = false
		r.gap("flow-budget", at.Span())
		return Value{}, nil, false
	}
	switch strings.ToLower(strings.TrimPrefix(name, `\`)) {
	case "assert", "extract", "compact", "get_defined_vars", "func_get_arg", "func_get_args", "func_num_args", "create_function":
		return Value{}, nil, false
	}
	proxy := *call
	proxy.Name = &syntax.Name{Value: `\` + strings.TrimPrefix(name, `\`)}
	proxy.Args = args
	if _, known := e.bind(&proxy, values); !known {
		return Value{}, nil, false
	}
	// Normalize fixed runtime roles for contracts that inspect flags or
	// omissions. Values are snapshots; synthetic source nodes establish no new
	// accepted-argument facts about variables at their old producer positions.
	if ordered, known := e.types.Native().BuiltinArguments(&proxy); known {
		function := e.types.ResolveFunction(&proxy)
		bound, seen, _ := e.bindValues(args, function.Params, values, false)
		normalized := &syntax.ArgList{}
		var normalizedValues []Value
		missing := false
		for i, given := range seen {
			if !given {
				missing = true
				continue
			}
			if bound[i].Expr == nil {
				if captured, known := r.values[ordered[i]]; known && captured.Complete {
					bound[i].Expr = captured.Expr
					// A sealed callback pack holds the resource captured at creation,
					// even when its source variable later names another stream.
					if i == 0 && strings.EqualFold(strings.TrimPrefix(name, `\`), "fclose") && r.allocations[captured.Identity] && captured.Invalidated == 0 && !captured.replaced {
						bound[i].Identity = captured.Identity
					}
				}
			}
			arg := &syntax.Arg{Value: ordered[i]}
			if missing {
				arg.Name = &syntax.Identifier{Value: strings.TrimPrefix(function.Params[i].Name, "$")}
			}
			normalized.Args = append(normalized.Args, arg)
			normalizedValues = append(normalizedValues, bound[i])
		}
		args, values, proxy.Args = normalized, normalizedValues, normalized
	}
	before := clone(f)
	firstCall := len(r.calls)
	var written []*syntax.Arg
	if args != nil {
		for _, expr := range args.Args {
			written = append(written, expr.(*syntax.Arg))
		}
	}
	e.invocationDepth++
	value := e.finishCall(&proxy, args, &evaluatedCall{prepared: true, followed: followTyped, closedFollowed: followTyped, values: values, written: written}, f, r)
	e.invocationDepth--
	var sinks []Sink
	for _, called := range r.calls[firstCall:] {
		sinks = appendSinks(sinks, e.nativeSinks(called))
		sinks = appendSinks(sinks, r.closureSinks[called.Node])
		delete(r.callIndex, called.Node)
		delete(r.closureSinks, called.Node)
	}
	r.calls = r.calls[:firstCall]
	if r.closureSinks == nil {
		r.closureSinks = map[syntax.Node][]Sink{}
	}
	r.closureSinks[at] = appendSinks(r.closureSinks[at], sinks)
	// Keep an indirect close pending until its exact result proves success.
	// Pending states are not closed-state proofs on failure or unchecked paths.
	for id, state := range f.states {
		if state.Span == at.Span() && !state.closeChecked && (state.Operation == "fclose" || state.Operation == "pclose" || state.Operation == "curl_close") {
			state.Operation = "@indirect:" + state.Operation
			f.setState(id, state)
		}
	}
	states := map[uint32]State{}
	for id, state := range before.states {
		if next, exists := f.states[id]; !exists || next != state {
			states[id] = next
		}
	}
	for id, state := range f.states {
		if old, exists := before.states[id]; !exists || old != state {
			states[id] = state
		}
	}
	return value, states, value.Complete && r.complete
}
