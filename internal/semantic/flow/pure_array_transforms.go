package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) column(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 3 || c.env.types.PHP.Below(phpversion.PHP55) {
		return nil
	}
	args := c.arguments(node, []pureParam{{name: params[0].Name}, {name: params[1].Name}, {name: params[2].Name}}, params)
	if args == nil {
		return nil
	}
	result := pureNode(pureColumn, args[0], args[1], args[2])
	if result != nil {
		result.projects = true
	}
	return result
}

func (c *pureCompiler) arrayMerge(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 1 || !params[0].Variadic || params[0].ByRef || node.Args == nil || len(node.Args.Args) != 2 {
		return nil
	}
	for _, item := range node.Args.Args {
		arg, ok := item.(*syntax.Arg)
		if !ok || arg.Name != nil || arg.Unpack {
			return nil
		}
	}
	declared := []index.Param{params[0], params[0]}
	declared[0].Variadic, declared[1].Variadic = false, false
	declared[1].Name += "2"
	args := c.arguments(node, []pureParam{{name: declared[0].Name}, {name: declared[1].Name}}, declared)
	if args == nil {
		return nil
	}
	result := pureNode(pureMerge, args[0], args[1], nil)
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalArrayTransform(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok || !purePlainRecord(left) || len(left.elements.order) != len(left.elements.items) {
		return Value{}, false
	}
	if expr.kind == pureColumn {
		key, known := pureRecordKey(right)
		if !known {
			return Value{}, false
		}
		indexKey := ""
		if expr.other != nil {
			index, valid := e.evalPure(expr.other, slots, r, depth-1)
			var known bool
			indexKey, known = pureRecordKey(index)
			if !valid || !known {
				return Value{}, false
			}
		}
		result := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{}}}
		if expr.other != nil {
			result.elements.width = len(left.elements.order)
		}
		for _, rowKey := range left.elements.order {
			row := left.elements.items[rowKey]
			if !purePlainRecord(row) {
				return Value{}, false
			}
			value, present := row.elements.lookup(key)
			if !present || !pureReadOnlyInputs([]Value{value}, nil) {
				return Value{}, false
			}
			if expr.other == nil {
				if !appendPureEntry(&result, value, r) {
					return Value{}, false
				}
			} else {
				index, held := row.elements.lookup(indexKey)
				key, known := pureRecordKey(index)
				if !held || !known || len(index.Sources) != 0 {
					return Value{}, false
				}
				r.copies++
				if r.copies > MaxCopied || len(result.elements.items) >= pureLimit {
					return Value{}, false
				}
				putWinningEntry(result.elements, key, value)
			}
			r.operations++
			if r.operations > MaxTransfers {
				return Value{}, false
			}
		}
		if expr.other != nil {
			result = winningRecord(result.elements)
		}
		return result, result.Complete
	}
	if !purePlainRecord(right) || len(right.elements.order) != len(right.elements.items) {
		return Value{}, false
	}
	count := len(left.elements.order) + len(right.elements.order)
	r.copies += count
	r.operations += count
	if count > pureLimit || r.copies > MaxCopied || r.operations > MaxTransfers {
		return Value{}, false
	}
	if expr.kind == pureCombine {
		if len(left.elements.order) == 0 || len(left.elements.order) != len(right.elements.order) {
			return Value{}, false
		}
		set := &elementSet{whole: true, width: count, items: map[string]Value{}}
		for i, at := range left.elements.order {
			value := left.elements.items[at]
			key, known := pureRecordKey(value)
			if !known || len(value.Sources) != 0 {
				return Value{}, false
			}
			putWinningEntry(set, key, right.elements.items[right.elements.order[i]])
		}
		result := winningRecord(set)
		return result, result.Complete
	}
	set := &elementSet{whole: true, width: count, items: map[string]Value{}}
	next := 0
	for _, input := range []Value{left, right} {
		for _, key := range input.elements.order {
			value := input.elements.items[key]
			if expr.kind == pureMerge && pureKeyScalar(key).k == 'i' {
				key = strconv.Itoa(next)
				next++
			}
			putWinningEntry(set, key, value)
		}
	}
	result := winningRecord(set)
	return result, result.Complete
}

func winningRecord(set *elementSet) Value {
	set.width = len(set.items)
	result := Value{Complete: true, plain: true, deep: true}
	for _, key := range set.order {
		result = merge(result, set.items[key])
	}
	result.elements, result.plain, result.deep = set, true, true
	return result
}

// Keep only winning value provenance while retaining first insertion order.
func putWinningEntry(set *elementSet, key string, value Value) {
	if _, held := set.items[key]; !held {
		set.order = append(set.order, key)
	}
	set.put(key, scalarSnapshot(value))
}

func (c *pureCompiler) pairedArray(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	if len(params) != 2 {
		return nil
	}
	declared := append([]index.Param(nil), params...)
	kind := pureCombine
	if name == "array_replace" {
		if !declared[1].Variadic || node.Args == nil || len(node.Args.Args) != 2 {
			return nil
		}
		for _, item := range node.Args.Args {
			arg, ok := item.(*syntax.Arg)
			if !ok || arg.Name != nil || arg.Unpack {
				return nil
			}
		}
		declared[1].Variadic = false
		kind = pureArrayReplace
	}
	args := c.arguments(node, []pureParam{{name: declared[0].Name}, {name: declared[1].Name}}, declared)
	if args == nil {
		return nil
	}
	result := pureNode(kind, args[0], args[1], nil)
	if result != nil {
		result.projects = true
	}
	return result
}
