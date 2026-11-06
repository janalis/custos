package flow

import (
	"math"
	"math/bits"
	"strconv"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func (c *pureCompiler) arrayConstruction(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	var kind uint8
	size, required := 1, 1
	switch name {
	case "array_fill":
		kind, size, required = pureFill, 3, 3
	case "array_fill_keys":
		kind, size, required = pureFillKeys, 2, 2
	case "array_flip":
		kind = pureFlip
	case "array_chunk":
		kind, size, required = pureChunk, 3, 2
	case "array_sum":
		kind = pureSum
	default:
		return nil
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
	if size == 1 {
		child := args[0]
		if child.cost >= pureLimit || child.depth >= MaxDepth {
			return nil
		}
		return &pureExpr{kind: kind, left: child, projects: true, cost: child.cost + 1, depth: child.depth + 1}
	}
	var other *pureExpr
	if size == 3 {
		other = args[2]
		if other == nil {
			other = pureConstant(constantScalar{k: 'b'})
		}
	}
	result := pureNode(kind, args[0], args[1], other)
	if result != nil {
		result.projects = true
	}
	return result
}

// Reserve the combined construction before publishing any element. Each table
// also reserves its width before depth checks, including duplicate-key inputs.
func pureArrayBudget(size int, r *scopeResult) bool {
	r.copies += size
	r.operations += size
	return size <= pureLimit && r.copies <= MaxCopied && r.operations <= MaxTransfers
}

func constructPureEntry(set *elementSet, key string, value Value) bool {
	if !pureReadOnlyInputs([]Value{value}, nil) || len(value.Sources) > MaxSources || value.elements != nil && value.elements.depth+bits.Len(uint(max(1, set.width))) > MaxDepth {
		return false
	}
	putWinningEntry(set, key, value)
	return true
}

func pureShapeInteger(value Value) (int64, bool) {
	scalar, known := knownScalar(value)
	return scalar.i, known && scalar.k == 'i' && len(value.Sources) == 0 && pureReadOnlyInputs([]Value{value}, nil)
}

func (e *Env) evalArrayConstruction(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if expr.kind == pureFill {
		return e.evalPureFill(expr, left, slots, r, depth)
	}
	if !purePlainRecord(left) || len(left.elements.order) != len(left.elements.items) {
		return Value{}, false
	}
	if expr.kind == pureChunk {
		return e.evalPureChunk(expr, left, slots, r, depth)
	}
	if expr.kind == pureSum {
		return evalPureSum(left, r)
	}
	var value Value
	if expr.kind == pureFillKeys {
		var ok bool
		value, ok = e.evalPure(expr.right, slots, r, depth-1)
		if !ok || !pureReadOnlyInputs([]Value{value}, nil) || len(value.Sources) > MaxSources {
			return Value{}, false
		}
	}
	size := len(left.elements.order)
	if !pureArrayBudget(size, r) {
		return Value{}, false
	}
	set := &elementSet{whole: true, width: size, items: map[string]Value{}}
	for _, at := range left.elements.order {
		entry := left.elements.items[at]
		key, known := pureRecordKey(entry)
		if !known || len(entry.Sources) != 0 || !pureReadOnlyInputs([]Value{entry}, nil) {
			return Value{}, false
		}
		item := value
		if expr.kind == pureFlip {
			item = scalarValue(pureKeyScalar(at))
		}
		if !constructPureEntry(set, key, item) {
			return Value{}, false
		}
	}
	result := winningRecord(set)
	return result, result.Complete
}

func (e *Env) evalPureFill(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	count, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	value, ok := e.evalPure(expr.other, slots, r, depth-1)
	start, sk := pureShapeInteger(left)
	size, nk := pureShapeInteger(count)
	if !ok || !sk || !nk || start < 0 || size <= 0 || size > pureLimit || start > math.MaxInt64-(size-1) {
		return Value{}, false
	}
	if !pureArrayBudget(int(size), r) {
		return Value{}, false
	}
	set := &elementSet{whole: true, width: int(size), items: map[string]Value{}}
	for i := int64(0); i < size; i++ {
		if !constructPureEntry(set, strconv.FormatInt(start+i, 10), value) {
			return Value{}, false
		}
	}
	result := winningRecord(set)
	return result, result.Complete
}

func (e *Env) evalPureChunk(expr *pureExpr, input Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	count, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	preserve, ok := e.evalPure(expr.other, slots, r, depth-1)
	flag, fk := knownScalar(preserve)
	size, sk := pureShapeInteger(count)
	if !ok || !sk || size <= 0 || !fk || flag.k != 'b' || !pureReadOnlyInputs([]Value{preserve}, nil) {
		return Value{}, false
	}
	length := len(input.elements.order)
	width := length
	if size < int64(width) {
		width = int(size)
	}
	chunks := 0
	if length != 0 {
		chunks = 1 + (length-1)/width
	}
	if !pureArrayBudget(length+chunks, r) {
		return Value{}, false
	}
	outer := &elementSet{whole: true, width: chunks, items: map[string]Value{}}
	for start := 0; start < length; start += width {
		end := min(start+width, length)
		inner := &elementSet{whole: true, width: end - start, items: map[string]Value{}}
		for at := start; at < end; at++ {
			key := input.elements.order[at]
			value := input.elements.items[key]
			if !flag.b {
				key = strconv.Itoa(at - start)
			}
			if !constructPureEntry(inner, key, value) {
				return Value{}, false
			}
		}
		if !constructPureEntry(outer, strconv.Itoa(start/width), winningRecord(inner)) {
			return Value{}, false
		}
	}
	result := winningRecord(outer)
	return result, result.Complete
}

func evalPureSum(input Value, r *scopeResult) (Value, bool) {
	if !pureArrayBudget(len(input.elements.order), r) {
		return Value{}, false
	}
	total := int64(0)
	var sources []Source
	for _, key := range input.elements.order {
		value := input.elements.items[key]
		scalar, known := knownScalar(value)
		if !known || scalar.k != 'i' || !pureReadOnlyInputs([]Value{value}, nil) || len(value.Sources) > MaxSources || scalar.i > 0 && total > math.MaxInt64-scalar.i || scalar.i < 0 && total < math.MinInt64-scalar.i {
			return Value{}, false
		}
		var complete bool
		sources, complete = union(sources, value.Sources)
		if !complete {
			return Value{}, false
		}
		total += scalar.i
	}
	result := scalarValue(constantScalar{k: 'i', i: total})
	result.Sources, result.Safe = sources, everyContext
	return result, true
}
