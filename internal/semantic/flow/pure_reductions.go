package flow

import (
	"maps"
	"strconv"

	"custos/internal/php/syntax"
)

// Recognize only an initialized carry, a single assignment, and its final
// return. No loop-local state or control flow escapes this contract.
func (c *pureCompiler) reduction(name string, initial *pureExpr, loop *syntax.Foreach, tail []syntax.Stmt) *pureExpr {
	if loop.ByRef || len(tail) != 1 {
		return nil
	}
	ret, ok := tail[0].(*syntax.Return)
	if !ok {
		return nil
	}
	returnedVariable, direct := ret.Expr.(*syntax.Variable)
	direct = direct && returnedVariable.NameExpr == nil && returnedVariable.Name == name
	final := *c
	if !direct {
		final.vars = append([]pureBinding(nil), c.vars...)
		final.extra = maps.Clone(c.extra)
		final.work = 0
		final.calls = nil
	}
	item, ok := loop.Value.(*syntax.Variable)
	if !ok || item.NameExpr != nil || item.Name == name || external(item.Name) || item.Name == "GLOBALS" || item.Name == "this" {
		return nil
	}
	body := loop.Body
	if block, ok := body.(*syntax.Block); ok {
		if len(block.Stmts) != 1 {
			return nil
		}
		body = block.Stmts[0]
	}
	var condition syntax.Expr
	if branch, filtered := body.(*syntax.If); filtered {
		if branch.Else != nil || len(branch.ElseIfs) != 0 {
			return nil
		}
		condition, body = branch.Cond, branch.Body
		if block, ok := body.(*syntax.Block); ok {
			if len(block.Stmts) != 1 {
				return nil
			}
			body = block.Stmts[0]
		}
	}
	stmt, ok := body.(*syntax.ExprStmt)
	if !ok {
		return nil
	}
	update, ok := stmt.Expr.(*syntax.Assign)
	if !ok || update.ByRef || update.Op.Kind != syntax.TEqual && update.Op.Kind != syntax.TPlusEqual && update.Op.Kind != syntax.TConcatEqual {
		return nil
	}
	build := false
	targetExpr := update.Var
	if dim, appendOnly := targetExpr.(*syntax.ArrayDimFetch); appendOnly && dim.Dim == nil && !dim.Curly && update.Op.Kind == syntax.TEqual && initial.kind == pureArray && len(initial.args) == 0 {
		targetExpr, build = dim.Var, true
	}
	target, ok := targetExpr.(*syntax.Variable)
	if !ok || target.NameExpr != nil || target.Name != name {
		return nil
	}
	if condition != nil && !build {
		return nil
	}
	keyName := ""
	if loop.Key != nil {
		key, ok := loop.Key.(*syntax.Variable)
		if !ok || key.NameExpr != nil || key.Name == name || key.Name == item.Name || external(key.Name) || key.Name == "GLOBALS" || key.Name == "this" {
			return nil
		}
		keyName = key.Name
	}
	input := c.expression(loop.Expr)
	slot := c.nextSlot
	c.bind(name, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	c.bind(item.Name, &pureExpr{kind: pureSlot, slot: slot + 1, cost: 1, depth: 1})
	if keyName != "" {
		c.bind(keyName, &pureExpr{kind: pureSlot, slot: slot + 2, cost: 1, depth: 1})
	}
	var guard *pureExpr
	if condition != nil {
		guard = c.expression(condition)
		if guard == nil {
			return nil
		}
	}
	value := c.expression(update.Value)
	if update.Op.Kind != syntax.TEqual {
		value = pureNode(pureBinary, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1}, value, nil)
		if value != nil {
			value.op = syntax.TPlus
			if update.Op.Kind == syntax.TConcatEqual {
				value.op = syntax.TDot
			}
		}
	}
	result := pureNode(pureReduce, input, initial, value)
	if result != nil {
		result.slot = slot
		if build {
			result.kind = pureBuild
		}
		if guard != nil {
			result.kind = pureFilterBuild
			result.args = []*pureExpr{guard}
			result.cost += guard.cost
			result.depth = max(result.depth, guard.depth+1)
			if result.cost > pureLimit || result.depth > MaxDepth {
				return nil
			}
		}
		if keyName != "" {
			result.op = syntax.TDoubleArrow
		}
		result.projects = true
	}
	if result == nil {
		return nil
	}
	if direct {
		return result
	}
	// A loop binding can be undefined for an empty input, or can overwrite an
	// original parameter. Only the carry and unchanged entry bindings survive.
	final.bind(item.Name, nil)
	if keyName != "" {
		final.bind(keyName, nil)
	}
	final.bind(name, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	returned := final.expression(ret.Expr)
	c.work += final.work
	c.calls = append(c.calls, final.calls...)
	if c.work > pureLimit {
		return nil
	}
	result = pureNode(pureLet, result, returned, nil)
	if result != nil {
		result.slot = slot
	}
	return result
}

func (e *Env) evalPureReduction(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	input, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok || !input.Complete || input.replaced || input.Invalidated != 0 || !input.plain || !input.deep || input.elements == nil || !input.elements.whole || input.elements.optional || input.elements.hasResources || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	carry, ok := e.evalPure(expr.right, slots, r, depth-1)
	if _, scalar := knownScalar(carry); !ok || !scalar && !((expr.kind == pureBuild || expr.kind == pureFilterBuild) && purePlainRecord(carry)) {
		return Value{}, false
	}
	width := 2
	if expr.op == syntax.TDoubleArrow {
		width = 3
	}
	r.copies += max(len(slots), expr.slot+width)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+width))
	copy(bound, slots)
	for _, key := range input.elements.order {
		item := input.elements.items[key]
		if _, scalar := knownScalar(item); !scalar && !((expr.kind == pureBuild || expr.kind == pureFilterBuild) && pureReadOnlyInputs([]Value{item}, nil)) {
			return Value{}, false
		}
		bound[expr.slot], bound[expr.slot+1] = carry, item
		if expr.op == syntax.TDoubleArrow {
			keyValue := constantScalar{k: 's', s: key}
			if number, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(number, 10) == key {
				keyValue = constantScalar{k: 'i', i: number}
			}
			bound[expr.slot+2] = scalarValue(keyValue)
		}

		if expr.kind == pureFilterBuild {
			selected, valid := e.evalPure(expr.args[0], bound, r, depth-1)
			predicate, known := knownScalar(selected)
			if !valid || !known {
				return Value{}, false
			}
			if !predicate.truth() {
				continue
			}
		}
		if expr.kind == pureBuild || expr.kind == pureFilterBuild {
			entry, valid := e.evalPure(expr.other, bound, r, depth-1)
			if !valid || !pureReadOnlyInputs([]Value{entry}, nil) || !appendPureEntry(&carry, entry, r) {
				return Value{}, false
			}
			continue
		}
		carry, ok = e.evalPure(expr.other, bound, r, depth-1)
		if _, scalar := knownScalar(carry); !ok || !scalar {
			return Value{}, false
		}
	}
	return carry, true
}

func (e *Env) evalPureLet(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	value, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	r.copies += max(len(slots), expr.slot+1)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+1))
	copy(bound, slots)
	bound[expr.slot] = value
	return e.evalPure(expr.right, bound, r, depth-1)
}
