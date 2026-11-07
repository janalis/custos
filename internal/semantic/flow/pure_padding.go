package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// padding compiles array_pad and integer range construction.
func (c *pureCompiler) padding(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	kind, size, required := pureArrayPad, 3, 3
	if name == "range" {
		kind, required = pureRange, 2
	}
	if len(params) != size {
		return nil
	}
	roles := make([]pureParam, size)
	for i, p := range params {
		roles[i].name = p.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil {
		return nil
	}
	for _, arg := range args[:required] {
		if arg == nil {
			return nil
		}
	}
	// An omitted step stays nil: only then may the bounds be equal.
	result := pureNode(kind, args[0], args[1], args[2])
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalPadding(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	right, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	if expr.kind == pureRange {
		return e.evalPureRange(expr, left, right, slots, r, depth)
	}
	value, ok := e.evalPure(expr.other, slots, r, depth-1)
	size, sk := pureShapeInteger(right)
	if !ok || !sk || !purePlainRecord(left) || len(left.elements.order) != len(left.elements.items) || !pureReadOnlyInputs([]Value{value}, nil) || len(value.Sources) > MaxSources {
		return Value{}, false
	}
	length := len(left.elements.order)
	target := size
	if target < 0 {
		target = -target
	}
	if target <= int64(length) {
		// No padding: PHP returns the array with its keys unchanged.
		return left, true
	}
	if target > pureLimit || !pureArrayBudget(int(target), r) {
		return Value{}, false
	}
	pads := int(target) - length
	set := &elementSet{whole: true, width: int(target), items: map[string]Value{}}
	next := 0
	add := func(value Value) bool {
		next++
		return constructPureEntry(set, strconv.Itoa(next-1), value)
	}
	padAll := func() bool {
		for range pads {
			if !add(value) {
				return false
			}
		}
		return true
	}
	if size < 0 && !padAll() {
		return Value{}, false
	}
	for _, key := range left.elements.order {
		entry := left.elements.items[key]
		if pureKeyScalar(key).k == 's' {
			if !constructPureEntry(set, key, entry) {
				return Value{}, false
			}
		} else if !add(entry) {
			return Value{}, false
		}
	}
	if size > 0 && !padAll() {
		return Value{}, false
	}
	result := winningRecord(set)
	return result, result.Complete
}

func (e *Env) evalPureRange(expr *pureExpr, left, right Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	start, sk := pureShapeInteger(left)
	end, ek := pureShapeInteger(right)
	if !sk || !ek {
		return Value{}, false
	}
	// Bounding both ends first keeps the subtractions below from overflowing;
	// a span beyond the construction budget is then rejected.
	if start < -1<<62 || end < -1<<62 || start > 1<<62 || end > 1<<62 || start-end > pureLimit || end-start > pureLimit {
		return Value{}, false
	}
	span := end - start
	if span < 0 {
		span = -span
	}
	step := int64(1)
	if expr.other != nil {
		given, ok := e.evalPure(expr.other, slots, r, depth-1)
		var known bool
		step, known = pureShapeInteger(given)
		// Zero, negative and oversized steps fail differently across versions.
		if !ok || !known || step < 1 || step > span {
			return Value{}, false
		}
	}
	count := span/step + 1
	if !pureArrayBudget(int(count), r) {
		return Value{}, false
	}
	direction := int64(1)
	if end < start {
		direction = -1
	}
	set := &elementSet{whole: true, width: int(count), items: map[string]Value{}}
	// Plain integers always fit: no entry can exceed the depth or source bounds.
	for i := int64(0); i < count; i++ {
		putWinningEntry(set, strconv.FormatInt(i, 10), scalarValue(constantScalar{k: 'i', i: start + direction*i*step}))
	}
	result := winningRecord(set)
	return result, result.Complete
}
