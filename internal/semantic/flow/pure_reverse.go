package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) reverse(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 2 {
		return nil
	}
	args := c.arguments(node, []pureParam{{name: params[0].Name}, {name: params[1].Name}}, params)
	if args == nil {
		return nil
	}
	if args[1] == nil {
		args[1] = pureConstant(constantScalar{k: 'b'})
	}
	result := pureNode(pureReverse, args[0], args[1], nil)
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalPureReverse(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	preserve, ok := e.evalPure(expr.right, slots, r, depth-1)
	flag, known := knownScalar(preserve)
	if !ok || !known || flag.k != 'b' || !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	size := len(input.elements.order)
	r.operations += size
	r.copies += size
	if size > pureLimit || r.operations > MaxTransfers || r.copies > MaxCopied {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: size}
	result := Value{Complete: true, plain: true, deep: true}
	next := 0
	for i := size - 1; i >= 0; i-- {
		key := input.elements.order[i]
		entry := input.elements.items[key]
		if !flag.b && pureKeyScalar(key).k == 'i' {
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
