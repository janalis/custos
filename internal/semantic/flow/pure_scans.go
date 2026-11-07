package flow

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// query compiles default-mode array_unique, strict array_search, checked
// integer array_product and byte-level strrev and substr_count. Optional
// modes (unique flags, search looseness, count offsets) establish nothing.
func (c *pureCompiler) query(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	var kind uint8
	size, required := 1, 1
	switch name {
	case "array_unique":
		kind, size = pureUnique, 2
	case "array_search":
		kind, size, required = pureSearch, 3, 3
	case "array_product":
		kind = pureProduct
	case "strrev":
		kind = pureTextScan
	default:
		kind, size, required = pureTextScan, 4, 2
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
	for i, arg := range args {
		if (i < required) != (arg != nil) {
			return nil
		}
	}
	mode := &constantScalar{k: 's', s: name}
	if required == 1 {
		child := args[0]
		if child.cost >= pureLimit || child.depth >= MaxDepth {
			return nil
		}
		return &pureExpr{kind: kind, literal: mode, left: child, projects: true, cost: child.cost + 1, depth: child.depth + 1}
	}
	var other *pureExpr
	if required == 3 {
		other = args[2]
	}
	result := pureNode(kind, args[0], args[1], other)
	if result != nil {
		result.literal, result.projects = mode, true
	}
	return result
}

// extreme compiles integer min and max: one array argument, or at least two
// positional arguments. Named, unpacked and by-reference forms are left out.
func (c *pureCompiler) extreme(node *syntax.FuncCall, name string) *pureExpr {
	if node.Args == nil || len(node.Args.Args) == 0 {
		return nil
	}
	expr := &pureExpr{kind: pureExtreme, literal: &constantScalar{k: 's', s: name}, projects: true, cost: 1, depth: 1}
	for i, item := range node.Args.Args {
		arg, ok := item.(*syntax.Arg)
		if !ok || arg.ByRef || arg.Unpack || arg.Name != nil {
			return nil
		}
		value := c.expression(arg.Value)
		if value == nil {
			return nil
		}
		if i == 0 {
			expr.left = value
		} else {
			expr.args = append(expr.args, value)
		}
		expr.cost += value.cost
		expr.depth = max(expr.depth, value.depth+1)
	}
	if expr.cost > pureLimit || expr.depth > MaxDepth {
		return nil
	}
	return expr
}

func (e *Env) evalQuery(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	switch expr.kind {
	case pureExtreme:
		return e.evalExtreme(expr, left, slots, r, depth)
	case pureTextScan:
		return e.evalTextScan(expr, left, slots, r, depth)
	case pureSearch:
		return e.evalSearch(expr, left, slots, r, depth)
	}
	if !purePlainRecord(left) || len(left.elements.order) != len(left.elements.items) || !pureArrayBudget(len(left.elements.order), r) {
		return Value{}, false
	}
	if expr.kind == pureProduct {
		return evalPureProduct(left)
	}
	// array_unique compares string forms and keeps the first key of each.
	set := &elementSet{whole: true, width: len(left.elements.order), items: map[string]Value{}}
	seen := map[string]bool{}
	for _, key := range left.elements.order {
		entry := left.elements.items[key]
		scalar, known := knownScalar(entry)
		if !known || scalar.k != 'i' && scalar.k != 's' {
			return Value{}, false
		}
		text := scalar.s
		if scalar.k == 'i' {
			text = strconv.FormatInt(scalar.i, 10)
		}
		if seen[text] {
			continue
		}
		seen[text] = true
		if !constructPureEntry(set, key, entry) {
			return Value{}, false
		}
	}
	result := winningRecord(set)
	return result, result.Complete
}

func (e *Env) evalSearch(expr *pureExpr, needle Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	haystack, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	strict, ok := e.evalPure(expr.other, slots, r, depth-1)
	flag, known := knownScalar(strict)
	if !ok || !known || flag.k != 'b' || !flag.b || !pureReadOnlyInputs([]Value{needle}, nil) || !purePlainRecord(haystack) || len(haystack.elements.order) != len(haystack.elements.items) || !pureArrayBudget(len(haystack.elements.order), r) {
		return Value{}, false
	}
	for _, key := range haystack.elements.order {
		equal, known := pureQueryIdentity(needle, haystack.elements.items[key], r, depth-1)
		if !known {
			return Value{}, false
		}
		if equal {
			return scalarValue(pureKeyScalar(key)), true
		}
	}
	return scalarValue(constantScalar{k: 'b'}), true
}

// evalExtreme picks the first smallest or largest of proven integers; the
// result is that entry, with its provenance.
func (e *Env) evalExtreme(expr *pureExpr, first Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	candidates := []Value{first}
	if len(expr.args) == 0 {
		if !purePlainRecord(first) || len(first.elements.order) != len(first.elements.items) || len(first.elements.order) == 0 || !pureArrayBudget(len(first.elements.order), r) {
			return Value{}, false
		}
		candidates = candidates[:0]
		for _, key := range first.elements.order {
			candidates = append(candidates, first.elements.items[key])
		}
	}
	for _, arg := range expr.args {
		value, ok := e.evalPure(arg, slots, r, depth-1)
		if !ok {
			return Value{}, false
		}
		candidates = append(candidates, value)
	}
	var best Value
	var top int64
	for i, candidate := range candidates {
		scalar, known := knownScalar(candidate)
		if !known || scalar.k != 'i' || !pureReadOnlyInputs([]Value{candidate}, nil) {
			return Value{}, false
		}
		if i == 0 || expr.literal.s == "max" && scalar.i > top || expr.literal.s == "min" && scalar.i < top {
			best, top = candidate, scalar.i
		}
	}
	return best, true
}

func evalPureProduct(input Value) (Value, bool) {
	product := int64(1)
	var sources []Source
	for _, key := range input.elements.order {
		value := input.elements.items[key]
		scalar, known := knownScalar(value)
		if !known || scalar.k != 'i' || !pureReadOnlyInputs([]Value{value}, nil) || len(value.Sources) > MaxSources {
			return Value{}, false
		}
		next, fits := multiplyChecked(product, scalar.i)
		var complete bool
		sources, complete = union(sources, value.Sources)
		if !fits || !complete {
			return Value{}, false
		}
		product = next
	}
	result := scalarValue(constantScalar{k: 'i', i: product})
	result.Sources, result.Safe = sources, everyContext
	return result, true
}

// multiplyChecked multiplies two integers, reporting whether the product
// fits: PHP turns an overflowing one into a float.
func multiplyChecked(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	if a == -1 || b == -1 {
		return a * b, a != math.MinInt64 && b != math.MinInt64
	}
	product := a * b
	return product, product/b == a
}

func (e *Env) evalTextScan(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	text, ok := pureBytes(left, pureTextBytes)
	if !ok {
		return Value{}, false
	}
	r.operations += len(text)
	if r.operations > MaxTransfers {
		return Value{}, false
	}
	if expr.literal.s == "strrev" {
		b := []byte(text)
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
		return pureSliceValue(left, string(b)), true
	}
	needle, ok := e.evalPure(expr.right, slots, r, depth-1)
	part, pk := pureBytes(needle, pureTrimBytes)
	if !ok || !pk || part == "" {
		return Value{}, false
	}
	sources, complete := union(left.Sources, needle.Sources)
	if !complete {
		return Value{}, false
	}
	result := scalarValue(constantScalar{k: 'i', i: int64(strings.Count(text, part))})
	result.Sources, result.Safe = sources, everyContext
	return result, true
}
