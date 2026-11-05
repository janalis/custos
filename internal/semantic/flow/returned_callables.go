package flow

import (
	"reflect"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// Returned closures publish only symbolic values, parameter metadata and sinks.
// Factory argument snapshots are supplied by each caller, never published.
type returnedCallable struct {
	forward           *forwardReturn
	argumentSensitive bool
	params            []index.Param
	defaults          []Value
	result            Value
	sinks             []Sink
}

type callableValue struct {
	contract *returnedCallable
	outer    []Value
}

func compactCallableValue(value Value) Value {
	value = scalarSnapshot(value)
	return Value{record: value.record, Complete: value.Complete, Sources: value.Sources, Safe: value.Safe, scalars: value.scalars, literal: value.literal}
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
		seeds[name] = compactReturnedValue(capture)
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
		contract.defaults = append(contract.defaults, compactReturnedValue(initial))
		seeds[param.Var.Name] = Value{Complete: true, Sources: []Source{{Kind: "parameter", Parameter: outer + i}}, Identity: param.Var.Span().Start + 1, plain: e.typed(param, false), deep: e.typed(param, true)}
	}
	previous, had := e.scopes[expr]
	e.invocationDepth++
	body := e.followSeeded(expr, nil, true, nil, seeds, nil, nil)
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
		result = body.values.at(arrow.Expr)
	} else {
		for i, value := range body.returns {
			if i == 0 {
				result = scalarSnapshot(value)
			} else {
				result = merge(result, scalarSnapshot(value))
			}
		}
	}
	if !result.Complete || result.targets != nil || result.returned != nil {
		return nil
	}
	result = completedReturn(expr, body, result)
	contract.forward = e.summarizeForward(expr, body, result)
	result = e.constrainCallbackReturn(expr, scalarSnapshot(result))
	if body.runs {
		result = eraseRecordResources(result)
	} else {
		result = recordContracts(result, body, nil)
	}
	contract.argumentSensitive = body.argumentSensitive || body.runs
	contract.result = compactReturnedValue(result)
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
		outer[i] = compactReturnedValue(value)
	}
	return &callableValue{contract: contract, outer: outer}
}

func (e *Env) invokeReturned(at syntax.Node, target *callableValue, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	contract := target.contract
	r.operations += len(contract.params) + len(contract.sinks) + len(target.outer) + 1
	r.copies += len(contract.params) + len(target.outer)
	if r.operations > MaxTransfers || r.copies > MaxCopied {
		r.complete = false
		r.gap("flow-budget", at.Span())
		return Value{}, nil, false
	}
	list, values, _ = userArguments(list, contract.params, values, !contract.argumentSensitive)
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
	forwarded, states, ok := e.forwardedReturn(at, contract.forward, bound, f, r)
	if ok {
		result = forwarded
	} else {
		result, states = e.instantiateRecord(at, result, args, f, r)
	}
	if r.closureSinks == nil {
		r.closureSinks = map[syntax.Node][]Sink{}
	}
	for _, sink := range contract.sinks {
		sink.Argument = substitute(sink.Argument, args, true)
		r.closureSinks[at] = appendSinks(r.closureSinks[at], []Sink{sink})
	}
	return result, states, true
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
		joined.outer[i] = compactReturnedValue(merge(value, b.outer[i]))
	}
	return joined
}

func compactReturnedValue(value Value) Value {
	out := compactCallableValue(value)
	fuel := MaxSources
	out.elements = compactElements(value.elements, MaxDepth, &fuel)
	return out
}

// Declared scalar bindings can coerce constants before a predicate observes
// them. In particular, negative floating zero becomes the truthy string "-0".
// Retain literal truth only for exact members or proven primitive conversions.
func (e *Env) constrainCallbackReturn(scope syntax.Node, value Value) Value {
	return e.constrainDeclaredValue(returnDeclaration(scope), value, !syntax.StrictTypes(e.file))
}

func (e *Env) constrainDeclaredValue(declared syntax.Expr, value Value, weak bool) Value {
	if declared == nil {
		return value
	}
	return e.constrainType(types.FromNode(declared, nil), value, weak)
}

func (e *Env) constrainType(result types.Type, value Value, weak bool) Value {
	if result.IsUnknown() || result.Has("mixed") {
		return value
	}
	if !result.Has("array") && !result.Has("iterable") {
		value.elements = nil
	}
	scalar, known := knownScalar(value)
	if known {
		name := "null"
		switch scalar.k {
		case 'i':
			name = "int"
		case 'f':
			name = "float"
		case 's':
			name = "string"
		case 'b':
			name = "bool"
		}
		if !result.Has(name) && !(name == "null" && result.Has("void")) {
			if len(result.Atoms()) == 1 && e.types.PHP.AtLeast(phpversion.PHP70) {
				if converted, valid := convertScalar(scalar, result.Atoms()[0], weak); valid {
					value.Expr = scalarExpression(converted)
					literal := converted
					value.literal = &literal
					value.interval, value.origins = nil, nil
					value.Identity, value.storedIdentity = 0, 0
					value.record = nil
					value.NonFalse, value.NonNull, value.LengthKnown = false, false, false
					return value
				}
			}
			value.literal = nil
			value.Expr, value.interval, value.origins = nil, nil, nil
			value.elements = nil
			value.Complete, value.NonFalse, value.NonNull, value.LengthKnown = false, false, false, false
		}
	}
	return value
}
