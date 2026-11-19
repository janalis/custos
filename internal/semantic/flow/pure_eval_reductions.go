package flow

import "github.com/janalis/custos/internal/php/syntax"

func (e *Env) evalPureReduction(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	input, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok || !input.Complete || input.replaced() || input.Invalidated != 0 || !input.plain || !input.deep || input.elements() == nil || !input.elements().whole || input.elements().optional || input.elements().hasResources || len(input.elements().order) != len(input.elements().items) {
		return Value{}, false
	}
	carry, ok := e.evalPure(expr.right, slots, r, depth-1)
	if _, scalar := knownScalar(carry); !ok || !scalar && !((expr.kind == pureBuild || expr.kind == pureFilterBuild) && purePlainRecord(carry)) {
		return Value{}, false
	}
	width := 2
	if len(expr.destruct) != 0 {
		width = 3 + len(expr.destruct)
	}
	if expr.op == syntax.TDoubleArrow {
		width = max(width, 3)
	}
	if expr.prepared {
		width = 4 + len(expr.destruct)
	}
	r.copies += max(len(slots), expr.slot+width)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+width))
	copy(bound, slots)
	for _, key := range input.elements().order {
		item := input.elements().items[key]
		if _, scalar := knownScalar(item); !scalar && !((expr.kind == pureBuild || expr.kind == pureFilterBuild || len(expr.destruct) != 0 || expr.prepared) && pureReadOnlyInputs([]Value{item}, nil)) {
			return Value{}, false
		}
		bound[expr.slot], bound[expr.slot+1] = carry, item
		if len(expr.destruct) != 0 && !bindPureFields(expr.destruct, item, bound, expr.slot, r) {
			return Value{}, false
		}
		if expr.op == syntax.TDoubleArrow {
			bound[expr.slot+2] = scalarValue(pureKeyScalar(key))
		}

		if expr.prepared {
			value, valid := e.evalPure(expr.args[0], bound, r, depth-1)
			if !valid || !pureReadOnlyInputs([]Value{value}, nil) {
				return Value{}, false
			}
			bound[expr.slot+3+len(expr.destruct)] = value
		}
		if expr.kind == pureFilterBuild || expr.kind == pureGuardReduce {
			selected, valid := e.evalPure(expr.args[0], bound, r, depth-1)
			truth, known := pureTruthValue(selected)
			if !valid || !known {
				return Value{}, false
			}
			if !truth {
				if expr.stopGuard {
					break
				}
				continue
			}
		}
		if expr.kind == pureBuild || expr.kind == pureFilterBuild {
			key, keyed := "", len(expr.args) > 1
			if keyed {
				value, ok := e.evalPure(expr.args[1], bound, r, depth-1)
				var known bool
				key, known = pureRecordKey(value)
				if !ok || !known {
					return Value{}, false
				}
			}
			entry, valid := e.evalPure(expr.other, bound, r, depth-1)
			if !valid || !pureReadOnlyInputs([]Value{entry}, nil) {
				return Value{}, false
			}
			if keyed {
				if !putPureKey(&carry, key, entry, r) {
					return Value{}, false
				}
			} else if !appendPureEntry(&carry, entry, r) {
				return Value{}, false
			}
			continue
		}
		carry, ok = e.evalPure(expr.other, bound, r, depth-1)
		if _, scalar := knownScalar(carry); !ok || !scalar {
			return Value{}, false
		}
	}
	return carry, true
}

func (e *Env) evalPureLet(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	value, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	r.copies += max(len(slots), expr.slot+1)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+1))
	copy(bound, slots)
	bound[expr.slot] = value
	return e.evalPure(expr.right, bound, r, depth-1)
}
