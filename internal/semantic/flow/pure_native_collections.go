package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// Bind against complete runtime-role metadata before selecting an audited form.
func (c *pureCompiler) collectionNative(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	size := 1
	if name == "array_keys" || name == "substr" {
		size = 3
	}
	if len(params) != size {
		return nil
	}
	if name == "array_keys" {
		return c.filteredKeys(node, params)
	}
	roles := make([]pureParam, size)
	for i, param := range params {
		roles[i].name = param.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil || args[0] == nil {
		return nil
	}
	if name == "substr" {
		result := pureNode(pureSubstring, args[0], args[1], args[2])
		if result != nil {
			result.projects = true
			result.weakArguments = c.env.types.PHP.Below(phpversion.PHP80)
			if c.env.types.PHP.Below(phpversion.PHP70) {
				result.slot = 1
			}
		}
		return result
	}
	child := args[0]
	if child.cost >= pureLimit || child.depth >= MaxDepth {
		return nil
	}
	kind, selector := pureOrderedArray, 0
	switch name {
	case "array_key_first":
		selector = 2
	case "array_key_last":
		selector = 3
	case "is_string":
		kind = pureTypeTest
	case "is_int":
		kind, selector = pureTypeTest, 1
	case "is_array":
		kind, selector = pureTypeTest, 2
	}
	return &pureExpr{kind: kind, slot: selector, left: child, projects: true, cost: child.cost + 1, depth: child.depth + 1}
}

func pureKeyScalar(key string) constantScalar {
	if number, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(number, 10) == key {
		return constantScalar{k: 'i', i: number}
	}
	return constantScalar{k: 's', s: key}
}

func (e *Env) evalCollectionNative(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if expr.kind == pureSubstring {
		return e.evalPureSubstring(expr, left, slots, r, depth)
	}
	if expr.kind == pureTypeTest {
		if !pureReadOnlyInputs([]Value{left}, nil) {
			return Value{}, false
		}
		scalar, known := knownScalar(left)
		record := purePlainRecord(left)
		if !known && !record {
			return Value{}, false
		}
		found := known && (expr.slot == 0 && scalar.k == 's' || expr.slot == 1 && scalar.k == 'i') || expr.slot == 2 && record
		return scalarValue(constantScalar{k: 'b', b: found}), true
	}
	if !purePlainRecord(left) || len(left.elements.order) != len(left.elements.items) {
		return Value{}, false
	}
	order := left.elements.order
	if expr.slot >= 2 {
		if len(order) == 0 {
			return scalarValue(constantScalar{k: 'n'}), true
		}
		at := 0
		if expr.slot == 3 {
			at = len(order) - 1
		}
		return scalarValue(pureKeyScalar(order[at])), true
	}
	r.copies += len(order)
	r.operations += len(order)
	if len(order) > pureLimit || r.copies > MaxCopied || r.operations > MaxTransfers {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: len(order)}
	result := Value{Complete: true, plain: true, deep: true}
	for i, key := range order {
		entry := scalarSnapshot(left.elements.items[key])
		if expr.slot == 1 {
			entry = scalarValue(pureKeyScalar(key))
		}
		text := strconv.Itoa(i)
		set.put(text, entry)
		set.order = append(set.order, text)
		result = merge(result, entry)
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}

// Normalize with int64 arithmetic before converting to machine indices. The
// input length is small, so even extreme negative offsets/lengths cannot wrap.
func (e *Env) evalPureSubstring(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	offset, ok := e.evalPure(expr.right, slots, r, depth-1)
	start, integer := knownScalar(offset)
	var length constantScalar
	if expr.other != nil {
		value, valid := e.evalPure(expr.other, slots, r, depth-1)
		var known bool
		length, known = knownScalar(value)
		if !valid || !known || length.k != 'i' {
			return Value{}, false
		}
	}
	input, stringKnown := knownScalar(left)
	if !ok || !integer || start.k != 'i' || !stringKnown || input.k != 's' || len(input.s) > 4096 || len(input.s) > MaxTransfers-r.operations {
		return Value{}, false
	}
	r.operations += len(input.s)
	size := int64(len(input.s))
	from := start.i
	if from < 0 {
		from = max(int64(0), size+from)
	}
	if from > size {
		if expr.weakArguments {
			return scalarValue(constantScalar{k: 'b'}), true
		}
		return pureSliceValue(left, ""), true
	}
	if expr.slot == 1 && from == size {
		return scalarValue(constantScalar{k: 'b'}), true
	}
	end := size
	if expr.other != nil {
		if length.i < 0 {
			end = size + length.i
		} else {
			end = from + min(length.i, size)
		}
	}
	if end < from {
		// Legacy negative offsets can truncate to empty even where a
		// positive offset reports false. A cut before byte zero still fails.
		if expr.weakArguments && (start.i >= 0 || end < 0) {
			return scalarValue(constantScalar{k: 'b'}), true
		}
		return pureSliceValue(left, ""), true
	}
	end = min(end, size)
	return pureSliceValue(left, input.s[int(from):int(end)]), true
}

// A slice carries its input provenance, but cannot inherit escaping proofs:
// cutting an encoded sequence can change the surrounding output context.
func pureSliceValue(input Value, text string) Value {
	value := scalarValue(constantScalar{k: 's', s: text})
	value.Sources = input.Sources
	return value
}
