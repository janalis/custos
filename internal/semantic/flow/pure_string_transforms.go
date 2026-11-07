package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) stringTransform(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	size := 2
	if name == "explode" {
		size = 3
	}
	if len(params) != size {
		return nil
	}
	roles := []pureParam{{name: params[0].Name}, {name: params[1].Name}}
	args := c.arguments(node, roles, params[:2])
	if args == nil {
		return nil
	}
	kind := pureJoin
	if name == "explode" {
		kind = pureSplit
	}
	result := pureNode(kind, args[0], args[1], nil)
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalStringTransform(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	separator, known := knownScalar(left)
	if !ok || !known || separator.k != 's' || len(separator.s) > 4096 {
		return Value{}, false
	}
	if expr.kind == pureSplit {
		return evalPureSplit(separator.s, right, r)
	}
	if !purePlainRecord(right) || len(right.elements.order) != len(right.elements.items) || len(right.elements.order) > pureLimit {
		return Value{}, false
	}
	parts := make([]string, 0, len(right.elements.order))
	var sources []Source
	size := 0
	if len(right.elements.order) > 1 {
		size = (len(right.elements.order) - 1) * len(separator.s)
		sources = left.Sources
	}
	for _, key := range right.elements.order {
		r.operations++
		value := right.elements.items[key]
		piece, known := knownScalar(value)
		if !known || piece.k != 's' || len(value.Sources) > MaxSources || len(sources) > MaxSources {
			return Value{}, false
		}
		size += len(piece.s)
		if size > 4096 {
			return Value{}, false
		}
		var complete bool
		sources, complete = union(sources, value.Sources)
		if !complete {
			return Value{}, false
		}
		parts = append(parts, piece.s)
	}
	r.operations += size + len(separator.s)
	r.copies += len(parts)
	if r.operations > MaxTransfers || r.copies > MaxCopied {
		return Value{}, false
	}
	result := scalarValue(constantScalar{k: 's', s: strings.Join(parts, separator.s)})
	result.Sources = sources
	return result, true
}

func evalPureSplit(separator string, input Value, r *scopeResult) (Value, bool) {
	text, known := knownScalar(input)
	if separator == "" || !known || text.k != 's' || len(text.s) > 4096 || len(input.Sources) > MaxSources {
		return Value{}, false
	}
	count := strings.Count(text.s, separator) + 1
	r.operations += 2 * (len(text.s) + len(separator))
	r.copies += count
	if count > pureLimit || r.operations > MaxTransfers || r.copies > MaxCopied {
		return Value{}, false
	}
	set := &elementSet{whole: true, items: map[string]Value{}, width: count}
	result := Value{Complete: true, plain: true, deep: true}
	for i, part := range strings.Split(text.s, separator) {
		key := strconv.Itoa(i)
		value := pureSliceValue(input, part)
		set.put(key, value)
		set.order = append(set.order, key)
		result = merge(result, value)
	}
	result.elements, result.plain, result.deep = set, true, true
	return result, true
}
