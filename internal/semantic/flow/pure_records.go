package flow

import (
	"maps"
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func pureTruthValue(value Value) (bool, bool) {
	if scalar, known := knownScalar(value); known {
		return scalar.truth(), true
	}
	if purePlainRecord(value) {
		return len(value.elements.items) != 0, true
	}
	return false, false
}

func (c *pureCompiler) empty(node *syntax.Empty) *pureExpr {
	if c.env.types.PHP.Below(phpversion.PHP55) {
		switch syntax.UnwrapParens(node.Expr).(type) {
		case *syntax.Variable, *syntax.ArrayDimFetch:
		default:
			return nil
		}
	}
	value := quietPure(c.expression(node.Expr))
	if value == nil || value.cost >= pureLimit || value.depth >= MaxDepth {
		return nil
	}
	return &pureExpr{kind: pureEmpty, left: value, projects: true, cost: value.cost + 1, depth: value.depth + 1}
}

func (e *Env) evalPureTruth(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	truth, known := pureTruthValue(left)
	if !known {
		return Value{}, false
	}
	if expr.kind == pureEmpty {
		return scalarValue(constantScalar{k: 'b', b: !truth}), true
	}
	if expr.kind == pureChoice {
		selected := expr.other
		if truth {
			selected = expr.right
		}
		return e.evalPure(selected, slots, r, depth-1)
	}
	and := expr.op == syntax.TBooleanAnd || expr.op == syntax.TAnd
	if truth != and {
		return scalarValue(constantScalar{k: 'b', b: !and}), true
	}
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	result, known := pureTruthValue(right)
	return scalarValue(constantScalar{k: 'b', b: result}), ok && known
}

func pureRecordKey(value Value) (string, bool) {
	scalar, known := knownScalar(value)
	if !known {
		return "", false
	}
	if scalar.k == 'i' {
		return strconv.FormatInt(scalar.i, 10), true
	}
	if scalar.k != 's' {
		return "", false
	}
	// Canonical numeric strings already share the internal spelling of integer
	// keys; other strings retain their spelling and therefore their PHP key type.
	return scalar.s, true
}

func putPureKey(value *Value, key string, entry Value, r *scopeResult) bool {
	set := value.elements
	_, held := set.items[key]
	r.copies += len(set.items) + 1
	if r.copies > MaxCopied || !held && len(set.items) >= pureLimit {
		return false
	}
	copied := *set
	copied.items, copied.order = maps.Clone(set.items), append([]string(nil), set.order...)
	if !held {
		copied.width++
		copied.order = append(copied.order, key)
	}
	copied.put(key, scalarSnapshot(entry))
	*value = merge(*value, entry)
	value.elements, value.plain, value.deep = &copied, true, true
	return true
}

func pureArrayUnion(left, right Value, r *scopeResult) (Value, bool) {
	if !purePlainRecord(left) || !purePlainRecord(right) || len(left.elements.order) != len(left.elements.items) || len(right.elements.order) != len(right.elements.items) {
		return Value{}, false
	}
	size := len(left.elements.items) + len(right.elements.items)
	r.copies += size
	r.operations += size
	if size > pureLimit || r.copies > MaxCopied || r.operations > MaxTransfers {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: size}
	for _, value := range []Value{left, right} {
		for _, key := range value.elements.order {
			if _, held := set.items[key]; held {
				continue
			}
			set.put(key, scalarSnapshot(value.elements.items[key]))
			set.order = append(set.order, key)
		}
	}
	result := merge(left, right)
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}
