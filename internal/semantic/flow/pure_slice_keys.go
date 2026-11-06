package flow

import (
	"slices"
	"strconv"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) slice(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 4 {
		return nil
	}
	roles := make([]pureParam, 4)
	for i, param := range params {
		roles[i].name = param.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil || args[0] == nil || args[1] == nil {
		return nil
	}
	if args[2] == nil {
		args[2] = pureConstant(constantScalar{k: 'n'})
	}
	if args[3] == nil {
		args[3] = pureConstant(constantScalar{k: 'b'})
	}
	result := pureNode(pureSlice, args[0], args[1], args[2])
	if result == nil {
		return nil
	}
	result.args = []*pureExpr{args[3]}
	result.cost += args[3].cost
	result.depth = max(result.depth, args[3].depth+1)
	result.projects = true
	if result.cost > pureLimit || result.depth > MaxDepth {
		return nil
	}
	return result
}

func (e *Env) evalPureSlice(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	offset, ok := e.evalPure(expr.right, slots, r, depth-1)
	length, lk := e.evalPure(expr.other, slots, r, depth-1)
	preserve, pk := e.evalPure(expr.args[0], slots, r, depth-1)
	start, sk := knownScalar(offset)
	limit, nk := knownScalar(length)
	keys, kk := knownScalar(preserve)
	if !ok || !lk || !pk || !sk || start.k != 'i' || !nk || limit.k != 'i' && limit.k != 'n' || !kk || keys.k != 'b' || !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	size := int64(len(input.elements.order))
	// Add only bounded nonnegative sizes to negative values, never negate MinInt64.
	first := min(start.i, size)
	if first < 0 {
		first = max(size+first, 0)
	}
	end := size
	if limit.k == 'i' {
		if limit.i < 0 {
			end = max(first, size+limit.i)
		} else {
			end = first + min(limit.i, size-first)
		}
	}
	count := int(end - first)
	r.copies += count
	r.operations += count
	if count > pureLimit || r.copies > MaxCopied || r.operations > MaxTransfers {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: count}
	result := Value{Complete: true, plain: true, deep: true}
	next := 0
	for _, key := range input.elements.order[first:end] {
		entry := input.elements.items[key]
		if !keys.b && pureKeyScalar(key).k == 'i' {
			key = strconv.Itoa(next)
			next++
		}
		set.put(key, scalarSnapshot(entry))
		set.order = append(set.order, key)
		result = merge(result, entry)
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}

func (c *pureCompiler) filteredKeys(node *syntax.FuncCall, params []index.Param) *pureExpr {
	// The stubs also expose an overload with a required search value. The
	// audited unfiltered native form omits it; preserve every supplied role.
	params = slices.Clone(params)
	params[1].Optional = true

	roles := make([]pureParam, len(params))
	for i, param := range params {
		roles[i].name = param.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil || args[0] == nil {
		return nil
	}
	if args[1] == nil {
		if args[2] != nil {
			return nil
		}
		child := args[0]
		if child.cost >= pureLimit || child.depth >= MaxDepth {
			return nil
		}
		return &pureExpr{kind: pureOrderedArray, left: child, slot: 1, projects: true, cost: child.cost + 1, depth: child.depth + 1}
	}
	if args[2] == nil {
		return nil
	}
	result := pureNode(pureFilteredKeys, args[0], args[1], args[2])
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalFilteredKeys(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	search, ok := e.evalPure(expr.right, slots, r, depth-1)
	strict, sk := e.evalPure(expr.other, slots, r, depth-1)
	flag, fk := knownScalar(strict)
	if !ok || !sk || !fk || flag.k != 'b' || !flag.b || !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) || !pureReadOnlyInputs([]Value{search}, nil) {
		return Value{}, false
	}
	result := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{}}}
	for _, key := range input.elements.order {
		r.operations++
		if r.operations > MaxTransfers {
			return Value{}, false
		}
		equal, known := pureQueryIdentity(search, input.elements.items[key], r, depth-1)
		if !known {
			return Value{}, false
		}
		if equal && !appendPureEntry(&result, scalarValue(pureKeyScalar(key)), r) {
			return Value{}, false
		}
	}
	return result, true
}
