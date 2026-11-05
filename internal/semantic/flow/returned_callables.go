package flow

import (
	"reflect"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// Returned closures publish only symbolic values, parameter metadata and sinks.
// Factory argument snapshots are supplied by each caller, never published.
type returnedCallable struct {
	params   []index.Param
	defaults []Value
	result   Value
	sinks    []Sink
}

type callableValue struct {
	contract *returnedCallable
	outer    []Value
}

func compactCallableValue(value Value) Value {
	return Value{Complete: value.Complete, Sources: value.Sources, Safe: value.Safe, scalars: value.scalars}
}

func (e *Env) summarizeReturned(value Value, owner *scopeResult, outer int) *returnedCallable {
	expr := syntax.UnwrapParens(value.Expr)
	captures, known := owner.closures[expr]
	if !known || !value.Complete || value.replaced || value.Invalidated != 0 || e.invocationDepth >= MaxDepth {
		return nil
	}
	params := syntax.VariableScopeParams(expr)
	if len(params) > 64 {
		return nil
	}
	contract := &returnedCallable{}
	seeds := map[string]Value{}
	for name, capture := range captures {
		if !capture.Complete || !capture.plain || !capture.deep || capture.replaced || capture.Invalidated != 0 {
			return nil
		}
		seeds[name] = compactCallableValue(capture)
	}
	lastRequired := -1
	for i, param := range params {
		if param.Default == nil {
			lastRequired = i
		}
	}
	for i, param := range params {
		if param.Var == nil || param.ByRef || param.Variadic {
			return nil
		}
		contract.params = append(contract.params, index.Param{Name: param.Var.Name, Optional: param.Default != nil && i > lastRequired})
		var initial Value
		if param.Default != nil {
			var valid bool
			initial, valid = e.closureDefault(param.Default, owner)
			if !valid {
				return nil
			}
		}
		contract.defaults = append(contract.defaults, compactCallableValue(initial))
		seeds[param.Var.Name] = Value{Complete: true, Sources: []Source{{Kind: "parameter", Parameter: outer + i}}, plain: e.typed(param, false), deep: e.typed(param, true)}
	}
	previous, had := e.scopes[expr]
	e.invocationDepth++
	body := e.followSeeded(expr, nil, true, nil, seeds, nil)
	e.invocationDepth--
	if had {
		e.scopes[expr] = previous
	} else {
		delete(e.scopes, expr)
	}
	owner.operations += body.operations
	owner.copies += body.copies
	if owner.reads != nil {
		addReads(owner.reads, body.reads)
	}
	if !body.complete || body.generator || len(body.statics) != 0 || len(body.globalNames) != 0 || owner.operations > MaxTransfers || owner.copies > MaxCopied {
		return nil
	}
	result := Value{Complete: true}
	if arrow, ok := expr.(*syntax.ArrowFunction); ok {
		result = body.values[arrow.Expr]
	} else {
		for i, value := range body.returns {
			if i == 0 {
				result = value
			} else {
				result = merge(result, value)
			}
		}
	}
	if !result.Complete || result.targets != nil || result.returned != nil {
		return nil
	}
	contract.result = compactCallableValue(result)
	for _, call := range body.calls {
		contract.sinks = appendSinks(contract.sinks, e.nativeSinks(call))
		contract.sinks = appendSinks(contract.sinks, e.Sinks(call.Node))
	}
	contract.sinks = appendSinks(contract.sinks, e.commands(syntax.VariableScopeBody(expr), body))
	return contract
}

func bindReturned(contract *returnedCallable, args []Value) *callableValue {
	if contract == nil {
		return nil
	}
	outer := make([]Value, len(args))
	for i, value := range args {
		outer[i] = compactCallableValue(value)
	}
	return &callableValue{contract: contract, outer: outer}
}

func (e *Env) invokeReturned(at syntax.Node, target *callableValue, list *syntax.ArgList, values []Value, r *scopeResult) (Value, map[uint32]State, bool) {
	contract := target.contract
	r.operations += len(contract.params) + len(contract.sinks) + len(target.outer) + 1
	r.copies += len(contract.params) + len(target.outer)
	if r.operations > MaxTransfers || r.copies > MaxCopied {
		r.complete = false
		r.gap("flow-budget", at.Span())
		return Value{}, nil, false
	}
	bound, seen, known := e.bindValues(list, contract.params, values, false)
	if !known {
		return Value{}, nil, false
	}
	for i, given := range seen {
		if !given {
			bound[i] = contract.defaults[i]
		}
	}
	args := append(append([]Value(nil), target.outer...), bound...)
	result := substitute(contract.result, args, true)
	if r.closureSinks == nil {
		r.closureSinks = map[syntax.Node][]Sink{}
	}
	for _, sink := range contract.sinks {
		sink.Argument = substitute(sink.Argument, args, true)
		r.closureSinks[at] = appendSinks(r.closureSinks[at], []Sink{sink})
	}
	return result, nil, true
}

func sameReturned(a, b *callableValue) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || !reflect.DeepEqual(a.contract, b.contract) || len(a.outer) != len(b.outer) {
		return false
	}
	for i, value := range a.outer {
		if !equal(value, b.outer[i]) {
			return false
		}
	}
	return true
}

func joinReturned(a, b *callableValue) *callableValue {
	if sameReturned(a, b) {
		return a
	}
	if a == nil || b == nil || !reflect.DeepEqual(a.contract, b.contract) || len(a.outer) != len(b.outer) {
		return nil
	}
	joined := &callableValue{contract: a.contract, outer: make([]Value, len(a.outer))}
	for i, value := range a.outer {
		joined.outer[i] = compactCallableValue(merge(value, b.outer[i]))
	}
	return joined
}
