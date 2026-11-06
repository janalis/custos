package flow

import (
	"maps"
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func pureNode(kind uint8, left, right, other *pureExpr) *pureExpr {
	if left == nil || right == nil || (kind == pureChoice || kind == pureReduce) && other == nil {
		return nil
	}
	cost, depth := 1+left.cost+right.cost, 1+max(left.depth, right.depth)
	if other != nil {
		cost += other.cost
		depth = max(depth, other.depth+1)
	}
	if cost > pureLimit || depth > MaxDepth {
		return nil
	}
	return &pureExpr{kind: kind, left: left, right: right, other: other, projects: kind == pureChoice || kind == pureShort || kind == pureIndex || kind == pureExists || kind == pureNonNull || left.projects || right.projects || other != nil && other.projects, cost: cost, depth: depth}
}

func (c *pureCompiler) statements(stmts []syntax.Stmt, depth int) *pureExpr {
	c.work++
	if depth == 0 || c.work > pureLimit {
		return nil
	}
	if len(stmts) == 0 {
		return pureConstant(constantScalar{k: 'n'})
	}
	switch n := stmts[0].(type) {
	case *syntax.Foreach:
		return c.find(n, stmts[1:])
	case *syntax.Return:
		if n.Expr == nil {
			return pureConstant(constantScalar{k: 'n'})
		}
		return c.expression(n.Expr)
	case *syntax.ExprStmt:
		assign, ok := n.Expr.(*syntax.Assign)
		if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual {
			return nil
		}
		variable, ok := assign.Var.(*syntax.Variable)
		if !ok || variable.NameExpr != nil || external(variable.Name) || variable.Name == "GLOBALS" || variable.Name == "this" {
			return nil
		}
		value := c.expression(assign.Value)
		if value == nil {
			return nil
		}
		c.bind(variable.Name, value)
		if len(stmts) > 1 {
			if loop, ok := stmts[1].(*syntax.Foreach); ok {
				return c.reduction(variable.Name, value, loop, stmts[2:])
			}
		}
		return pureNode(pureSequence, value, c.statements(stmts[1:], depth-1), nil)
	case *syntax.If:
		condition := c.expression(n.Cond)
		tail := stmts[1:]
		var otherwise syntax.Stmt
		if n.Else != nil {
			otherwise = n.Else.Body
		}
		no := c.branch(otherwise, tail, depth-1)
		for i := len(n.ElseIfs) - 1; i >= 0; i-- {
			branch := n.ElseIfs[i]
			no = pureNode(pureChoice, c.expression(branch.Cond), c.branch(branch.Body, tail, depth-1), no)
		}
		return pureNode(pureChoice, condition, c.branch(n.Body, tail, depth-1), no)
	}
	return nil
}

func (c *pureCompiler) branch(body syntax.Stmt, tail []syntax.Stmt, depth int) *pureExpr {
	if c.work >= pureLimit || depth == 0 {
		return nil
	}
	child := *c
	child.vars = append([]pureBinding(nil), c.vars...)
	child.extra = maps.Clone(c.extra)
	child.work = 0
	child.calls = nil
	var stmts []syntax.Stmt
	if block, ok := body.(*syntax.Block); ok {
		stmts = append(stmts, block.Stmts...)
	} else if body != nil {
		stmts = append(stmts, body)
	}
	stmts = append(stmts, tail...)
	result := child.statements(stmts, depth)
	c.work += child.work
	c.calls = append(c.calls, child.calls...)
	if c.work > pureLimit {
		return nil
	}
	return result
}

func (c *pureCompiler) array(array *syntax.Array) *pureExpr {
	if len(array.Items) > pureLimit || array.Short && c.env.types.PHP.Below(phpversion.PHP54) {
		return nil
	}
	for _, item := range array.Items {
		if item != nil && item.Unpack {
			return c.spreadArray(array)
		}
		if item != nil && item.Key != nil {
			fuel := 256
			if _, known := constantValue(item.Key, MaxDepth, &fuel); !known {
				return c.runtimeArray(array)
			}
		}
	}
	expr := &pureExpr{kind: pureArray, cost: 1, depth: 1}
	next := int64(0)
	seen := map[string]bool{}
	for _, item := range array.Items {
		if item == nil || item.ByRef || item.Unpack {
			return nil
		}
		key := strconv.FormatInt(next, 10)
		if item.Key != nil {
			fuel := 256
			scalar, constant := constantValue(item.Key, MaxDepth, &fuel)
			if !constant || scalar.k == 'f' {
				return nil
			}
			var known bool
			key, known = arrayKey(item.Key)
			if !known {
				return nil
			}
		}
		if seen[key] {
			return nil
		}
		seen[key] = true
		if number, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(number, 10) == key {
			// Negative implicit indexing changed in PHP 8.3. Keep it outside
			// these contracts until a version-specific sequence is proven.
			if number < 0 || number == math.MaxInt64 {
				return nil
			}
			next = max(next, number+1)
		}
		value := c.expression(item.Value)
		if value == nil {
			return nil
		}
		expr.args = append(expr.args, pureConstant(constantScalar{k: 's', s: key}), value)
		expr.cost += 1 + value.cost
		expr.depth = max(expr.depth, value.depth+1)
	}
	if expr.cost > pureLimit || expr.depth > MaxDepth {
		return nil
	}
	return expr
}

func (c *pureCompiler) native(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	if name == "implode" || name == "explode" {
		return c.stringTransform(node, name, params)
	}
	if name == "array_column" {
		return c.column(node, params)
	}
	if name == "array_merge" {
		return c.arrayMerge(node, params)
	}
	if name == "array_reduce" {
		return c.fold(node, params)
	}
	if name == "array_slice" {
		return c.slice(node, params)
	}
	if name == "array_map" || name == "array_filter" {
		return c.higher(node, name, params)
	}
	switch name {
	case "array_values", "array_keys", "array_key_first", "array_key_last", "is_string", "is_int", "is_array", "substr":
		return c.collectionNative(node, name, params)
	}

	if name == "str_starts_with" || name == "str_ends_with" || name == "str_contains" {
		return c.stringPredicate(node, name, params)
	}
	if name == "count" || name == "in_array" {
		return c.finiteNative(node, name, params)
	}
	if name == "array_key_exists" {
		return c.keyExists(node, params)
	}
	if name != "strlen" && name != "trim" || node.Args == nil || len(node.Args.Args) != 1 || len(params) == 0 {
		return nil
	}
	arg, ok := node.Args.Args[0].(*syntax.Arg)
	if !ok || arg.ByRef || arg.Unpack || arg.Name != nil && (c.env.types.PHP.Below(phpversion.PHP80) || arg.Name.Value != params[0].Name) {
		return nil
	}
	child := c.expression(arg.Value)
	if child == nil || child.cost >= pureLimit || child.depth >= MaxDepth {
		return nil
	}
	kind := pureLength
	if name == "trim" {
		kind = pureTrim
	}
	return &pureExpr{kind: kind, left: child, projects: child.projects, cost: child.cost + 1, depth: child.depth + 1}
}

// Ordinary calls keep their compact compiler binding; spreads use the same
// runtime role binder as value analysis, with scalar indices for symbolic terms.
func (c *pureCompiler) arguments(node syntax.Node, params []pureParam, declared []index.Param) []*pureExpr {
	// Every composed call requires the published contract and declaration to agree.
	// An editor can replace declaration metadata before an old snapshot expires.
	if len(declared) != len(params) {
		return nil
	}
	for i, param := range declared {
		if param.ByRef || param.Variadic || param.Name != params[i].name {
			return nil
		}
	}
	if arguments(node) != nil {
		for _, item := range arguments(node).Args {
			if arg, ok := item.(*syntax.Arg); ok && arg.Unpack {
				return c.spreadArguments(node, params, declared)
			}
		}
	}
	args := make([]*pureExpr, len(params))
	position := 0
	named := false
	if arguments(node) != nil {
		for _, item := range arguments(node).Args {
			arg, ok := item.(*syntax.Arg)
			if !ok || arg.ByRef {
				return nil
			}
			at := position
			if arg.Name != nil {
				if c.env.types.PHP.Below(phpversion.PHP80) {
					return nil
				}
				named, at = true, -1
				for i, param := range params {
					if param.name == arg.Name.Value {
						at = i
						break
					}
				}
			} else {
				if named {
					return nil
				}
				position++
			}
			if at < 0 || at >= len(args) || args[at] != nil {
				return nil
			}
			args[at] = c.expression(arg.Value)
			if args[at] == nil {
				return nil
			}
		}
	}
	return args
}

func (c *pureCompiler) spreadArguments(node syntax.Node, params []pureParam, declared []index.Param) []*pureExpr {
	var terms []*pureExpr
	marker := func(term *pureExpr) Value {
		index := len(terms)
		terms = append(terms, term)
		return scalarValue(constantScalar{k: 'i', i: int64(index)})
	}
	values := make([]Value, len(arguments(node).Args))
	for i, item := range arguments(node).Args {
		arg, ok := item.(*syntax.Arg)
		if !ok || arg.ByRef {
			return nil
		}
		term := c.expression(arg.Value)
		if term == nil {
			return nil
		}
		if !arg.Unpack {
			values[i] = marker(term)
			continue
		}
		if term.kind != pureArray {
			return nil
		}
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(term.args) / 2}
		for j := 0; j < len(term.args); j += 2 {
			key := term.args[j].literal.s
			set.put(key, marker(term.args[j+1]))
			set.order = append(set.order, key)
		}
		values[i] = Value{Complete: true, elements: set, plain: true, deep: true}
	}
	bound, seen, known := c.env.bindValues(arguments(node), declared, values, false)
	if !known {
		return nil
	}
	args := make([]*pureExpr, len(params))
	for i, given := range seen {
		if given {
			args[i] = terms[bound[i].literal.i]
		}
	}
	return args
}

func (e *Env) evalPureExtended(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if expr.kind == pureInline {
		return Value{}, false
	}
	if expr.kind == pureFold {
		return e.evalPureFold(expr, slots, r, depth)
	}
	if expr.kind == pureFind {
		return e.evalPureFind(expr, slots, r, depth)
	}
	if expr.kind == pureLet {
		return e.evalPureLet(expr, slots, r, depth)
	}
	if expr.kind == pureSpreadArray {
		return e.evalSpreadArray(expr, slots, r, depth)
	}
	if expr.kind == pureReduce || expr.kind == pureBuild || expr.kind == pureFilterBuild || expr.kind == pureGuardReduce {
		return e.evalPureReduction(expr, slots, r, depth)
	}
	if expr.kind == pureMatch {
		return e.evalPureMatch(expr, slots, r, depth)
	}
	if expr.kind == pureArray {
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(expr.args) / 2}
		value := Value{Complete: true, plain: true, deep: true}
		r.copies += len(expr.args) / 2
		if r.copies > MaxCopied {
			return Value{}, false
		}
		for i := 0; i < len(expr.args); i += 2 {
			entry, ok := e.evalPure(expr.args[i+1], slots, r, depth-1)
			if !ok || !entry.Complete || !entry.plain || !entry.deep || entry.Identity != 0 || entry.storedIdentity != 0 || entry.elements != nil && entry.elements.hasResources {
				return Value{}, false
			}
			key := expr.args[i].literal.s
			set.put(key, entry)
			set.order = append(set.order, key)
			value = merge(value, entry)
		}
		value.elements = set
		value.plain, value.deep = true, true
		return value, true
	}
	left, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	if expr.kind == pureJoin || expr.kind == pureSplit {
		return e.evalStringTransform(expr, left, slots, r, depth)
	}
	if expr.kind == pureColumn || expr.kind == pureMerge {
		return e.evalArrayTransform(expr, left, slots, r, depth)
	}
	if expr.kind == pureSlice {
		return e.evalPureSlice(expr, left, slots, r, depth)
	}
	if expr.kind == pureFilteredKeys {
		return e.evalFilteredKeys(expr, left, slots, r, depth)
	}
	if expr.kind == pureCast {
		return evalPureCast(expr, left)
	}
	if expr.kind == pureHigher {
		return e.evalPureHigher(expr, left, slots, r, depth)
	}
	if expr.kind == pureSequence {
		return e.evalPure(expr.right, slots, r, depth-1)
	}
	if expr.kind == pureIndex || expr.kind == pureQuietIndex || expr.kind == pureExists {
		// PHP evaluates offset expressions even on a quiet missing path.
		keyValue, known := e.evalPure(expr.right, slots, r, depth-1)
		key, scalar := knownScalar(keyValue)
		if !known || !scalar || key.k != 'i' && key.k != 's' {
			return Value{}, false
		}
		if scalar, known := knownScalar(left); expr.kind == pureQuietIndex && known && scalar.k == 'n' {
			return left, true
		}
		if !left.Complete || !left.plain || !left.deep || left.elements == nil || !left.elements.whole || left.elements.optional || left.elements.hasResources {
			return Value{}, false
		}

		text := key.s
		if key.k == 'i' {
			text = strconv.FormatInt(key.i, 10)
		}
		value, held := left.elements.lookup(text)
		if expr.kind == pureExists {
			return scalarValue(constantScalar{k: 'b', b: held}), true
		}
		if !held && expr.kind == pureQuietIndex {
			return scalarValue(constantScalar{k: 'n'}), true
		}
		return value, held && value.Complete && !value.replaced && value.Invalidated == 0
	}
	if expr.kind == pureNonNull || expr.kind == pureShort && expr.op == syntax.TCoalesce {
		nonnull, known := pureNonNullValue(left)
		if !known {
			return Value{}, false
		}
		if expr.kind == pureNonNull {
			return scalarValue(constantScalar{k: 'b', b: nonnull}), true
		}
		if nonnull {
			return left, true
		}
		return e.evalPure(expr.right, slots, r, depth-1)
	}
	if expr.kind == pureOrderedArray || expr.kind == pureTypeTest || expr.kind == pureSubstring {
		return e.evalCollectionNative(expr, left, slots, r, depth)
	}
	if expr.kind == pureStringPredicate {
		return e.evalStringPredicate(expr, left, slots, r, depth)
	}
	if expr.kind == pureCount || expr.kind == pureMember {
		return e.evalFiniteNative(expr, left, slots, r, depth)
	}
	if expr.kind == pureChoice || expr.kind == pureShort || expr.kind == pureEmpty {
		return e.evalPureTruth(expr, left, slots, r, depth)
	}
	scalar, known := knownScalar(left)
	if !known {
		return Value{}, false
	}
	switch expr.kind {
	case pureLength, pureTrim:
		if scalar.k != 's' || len(scalar.s) > 4096 {
			return Value{}, false
		}
		if expr.kind == pureLength {
			value := scalarValue(constantScalar{k: 'i', i: int64(len(scalar.s))})
			value.Sources = left.Sources
			value.Safe = everyContext
			return value, true
		}
		return pureSliceValue(left, strings.Trim(scalar.s, " \n\r\t\v\x00")), true
	}
	return Value{}, false
}
