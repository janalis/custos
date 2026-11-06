package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func purePlainRecord(value Value) bool {
	return value.Complete && !value.replaced && value.Invalidated == 0 && value.Identity == 0 && value.storedIdentity == 0 && value.plain && value.deep && value.elements != nil && value.elements.whole && !value.elements.optional && !value.elements.hasResources
}

func pureNonNullValue(value Value) (bool, bool) {
	if scalar, known := knownScalar(value); known {
		return scalar.k != 'n', true
	}
	return true, purePlainRecord(value)
}

func (c *pureCompiler) finiteNative(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	size := 2
	if name == "in_array" {
		size = 3
	}
	if len(params) != size {
		return nil
	}
	roles := make([]pureParam, size)
	for i, param := range params {
		roles[i].name = param.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil {
		return nil
	}
	if name == "count" {
		if args[1] == nil {
			args[1] = pureConstant(constantScalar{k: 'i'})
		}
		result := pureNode(pureCount, args[0], args[1], nil)
		if result != nil {
			result.projects = true
		}
		return result
	}
	if args[2] == nil {
		return nil
	}
	result := pureNode(pureMember, args[0], args[1], args[2])
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalFiniteNative(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	if expr.kind == pureCount {
		mode, known := knownScalar(right)
		if !known || mode.k != 'i' || mode.i != 0 || !purePlainRecord(left) {
			return Value{}, false
		}
		return scalarValue(constantScalar{k: 'i', i: int64(len(left.elements.items))}), true
	}
	needle, scalar := knownScalar(left)
	strict, ok := e.evalPure(expr.other, slots, r, depth-1)
	flag, known := knownScalar(strict)
	if !pureReadOnlyInputs([]Value{left}, nil) || !ok || !known || flag.k != 'b' || !flag.b || !purePlainRecord(right) || len(right.elements.order) != len(right.elements.items) {
		return Value{}, false
	}
	found, uncertain := false, !scalar && len(right.elements.items) != 0
	for _, key := range right.elements.order {
		entry := right.elements.items[key]
		r.operations++
		value, known := knownScalar(entry)
		if r.operations > MaxTransfers {
			return Value{}, false
		}
		if !known || !scalar {
			uncertain = true
			continue
		}
		equal, _ := constantBinary(syntax.TIsIdentical, needle, value)
		found = found || equal.b
	}
	if uncertain && !found {
		return Value{}, false
	}
	return scalarValue(constantScalar{k: 'b', b: found}), true
}
