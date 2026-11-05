package flow

import (
	"math"

	"custos/internal/php/syntax"
)

func scalarValue(value constantScalar) Value {
	literal := value
	return Value{Complete: true, literal: &literal, truth: scalarTruth(value), plain: true, deep: true}
}

func (e *Env) pureBind(contract *pureContract, args []Value, weak bool) ([]Value, bool) {
	if len(args) != len(contract.params) {
		return nil, false
	}
	bound := make([]Value, len(args))
	for i, param := range contract.params {
		value := args[i]
		if !value.Complete || value.replaced || value.Invalidated != 0 {
			return nil, false
		}
		if scalar, known := knownScalar(value); known {
			if !(param.nullable && scalar.k == 'n') {
				value = e.constrainType(param.typ, scalarValue(scalar), weak)
			}
		} else if !param.typ.IsUnknown() && !param.typ.Has("mixed") && (value.elements == nil || !recordForwardType(param.typ)) {
			return nil, false
		}
		if !value.Complete {
			return nil, false
		}
		bound[i] = value
	}
	return bound, true
}

func (e *Env) evalPure(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if depth == 0 || r.operations >= MaxTransfers {
		return Value{}, false
	}
	r.operations++
	switch expr.kind {
	case pureSlot:
		if expr.slot >= len(slots) {
			return Value{}, false
		}
		return slots[expr.slot], true
	case pureLiteral:
		return scalarValue(*expr.literal), true
	case pureCall:
		args := make([]Value, len(expr.args))
		r.copies += len(args)
		if r.copies > MaxCopied {
			return Value{}, false
		}
		for i, arg := range expr.args {
			value, known := e.evalPure(arg, slots, r, depth-1)
			if !known {
				return Value{}, false
			}
			args[i] = value
		}
		bound, known := e.pureBind(expr.call, args, expr.weakArguments)
		if !known {
			return Value{}, false
		}
		value, known := e.runPure(expr.call, bound, r, depth-1)
		if !known {
			return Value{}, false
		}
		value = e.constrainType(expr.call.result, value, expr.call.weakReturn)
		return value, value.Complete
	}
	left, known := e.evalPure(expr.left, slots, r, depth-1)
	if !known {
		return Value{}, false
	}
	a, known := knownScalar(left)
	if !known {
		return Value{}, false
	}
	var result constantScalar
	if expr.kind == pureUnary {
		switch expr.op {
		case syntax.TExclaim:
			result = constantScalar{k: 'b', b: !a.truth()}
		case syntax.TPlus:
			if a.k != 'i' && a.k != 'f' {
				return Value{}, false
			}
			result = a
		case syntax.TMinus:
			if a.k == 'i' && a.i != math.MinInt64 {
				result = constantScalar{k: 'i', i: -a.i}
			} else if a.k == 'f' {
				result = constantScalar{k: 'f', f: -a.f}
			} else {
				return Value{}, false
			}
		}
	} else {
		right, ok := e.evalPure(expr.right, slots, r, depth-1)
		if !ok {
			return Value{}, false
		}
		b, ok := knownScalar(right)
		if !ok {
			return Value{}, false
		}
		if expr.op == syntax.TDot && a.k == 's' && b.k == 's' {
			cost := len(a.s) + len(b.s)
			if cost > 4096 || cost > MaxTransfers-r.operations {
				return Value{}, false
			}
			r.operations += cost
		}
		result, known = constantBinary(expr.op, a, b)
		if !known {
			return Value{}, false
		}
	}
	return scalarValue(result), true
}

// Check discarded assignments too: evaluating an unused expression can warn or
// execute a conversion, so a return expression alone cannot prove purity.
func (e *Env) runPure(contract *pureContract, slots []Value, r *scopeResult, depth int) (Value, bool) {
	for _, step := range contract.steps {
		if _, known := e.evalPure(step, slots, r, depth); !known {
			return Value{}, false
		}
	}
	return e.evalPure(contract.expr, slots, r, depth)
}

// Scalar evaluation refines the existing provenance contract. Structural
// forwarding additionally needs an unchanged caller record and resource proofs.
func (e *Env) pureReturn(at syntax.Node, contract *pureContract, args, outer []Value, base Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	if contract == nil || len(outer) != contract.outer {
		return Value{}, nil, false
	}
	r.copies += len(args) + len(outer)
	if r.copies > MaxCopied {
		return Value{}, nil, false
	}
	bound, known := e.pureBind(contract, args, !syntax.StrictTypes(e.file) || e.internalScalarCallback(at))
	if !known {
		return Value{}, nil, false
	}
	slots := bound
	if len(outer) > 0 {
		slots = append(append([]Value(nil), outer...), bound...)
	}
	value, known := e.runPure(contract, slots, r, MaxDepth)
	if !known {
		return Value{}, nil, false
	}
	value = e.constrainType(contract.result, value, contract.weakReturn)
	if !value.Complete {
		return Value{}, nil, false
	}
	if scalar, known := knownScalar(value); known {
		literal := scalar
		base.literal, base.truth = &literal, scalarTruth(scalar)
		return base, nil, true
	}
	if value.elements == nil || !value.elements.whole || value.elements.optional || !wholeRecord(value.elements) {
		return Value{}, nil, false
	}
	selected := -1
	for i, arg := range args {
		if arg.elements == value.elements {
			selected = i
			break
		}
	}
	if selected < 0 {
		return Value{}, nil, false
	}
	for i, arg := range args {
		if i != selected && (!arg.plain || !arg.deep || arg.Identity != 0 || arg.elements != nil && arg.elements.hasResources) {
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
