package flow

import (
	"maps"
	"slices"
	"strconv"

	"custos/internal/php/syntax"
)

// Recognize only an initialized carry, a single assignment, and its final
// return. No loop-local state or control flow escapes this contract.
func (c *pureCompiler) reduction(name string, initial *pureExpr, loop *syntax.Foreach, tail []syntax.Stmt) *pureExpr {
	if loop.ByRef || len(tail) != 1 && len(tail) != 3 {
		return nil
	}
	pipeline := len(tail) == 3
	ret, ok := tail[len(tail)-1].(*syntax.Return)
	if !ok {
		return nil
	}
	returnedVariable, direct := ret.Expr.(*syntax.Variable)
	direct = direct && !pipeline && returnedVariable.NameExpr == nil && returnedVariable.Name == name
	final := *c
	if !direct {
		final.vars = append([]pureBinding(nil), c.vars...)
		final.extra = maps.Clone(c.extra)
		final.work = 0
		final.calls = nil
	}
	fields, destructured := c.loopFields(loop.Value, name)
	if fields == nil {
		return nil
	}
	overlaps := func(value string) bool {
		return slices.ContainsFunc(fields, func(field pureField) bool { return field.name == value })
	}
	body := loop.Body
	var skip syntax.Expr
	stop := false
	if block, ok := body.(*syntax.Block); ok && len(block.Stmts) == 2 {
		branch, ok := block.Stmts[0].(*syntax.If)
		if !ok || branch.Else != nil || len(branch.ElseIfs) != 0 {
			return nil
		}
		var levelExpr syntax.Expr
		switch transfer := singlePureStatement(branch.Body).(type) {
		case *syntax.Continue:
			levelExpr = transfer.Num
		case *syntax.Break:
			levelExpr, stop = transfer.Num, true
		default:
			return nil
		}
		if levelExpr != nil {
			if literal, ok := syntax.UnwrapParens(levelExpr).(*syntax.Literal); !ok || literal.LitKind != syntax.LitInt {
				return nil
			}
			fuel := 256
			level, known := constantValue(levelExpr, MaxDepth, &fuel)
			if !known || level.k != 'i' || level.i != 1 {
				return nil
			}
		}
		if _, ok := block.Stmts[1].(*syntax.ExprStmt); !ok {
			return nil
		}
		skip, body = branch.Cond, block.Stmts[1]
	}
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
	var targetKey syntax.Expr
	targetExpr := update.Var
	if dim, appendOnly := targetExpr.(*syntax.ArrayDimFetch); appendOnly && !dim.Curly && update.Op.Kind == syntax.TEqual && initial.kind == pureArray && len(initial.args) == 0 {
		targetExpr, build, targetKey = dim.Var, true, dim.Dim
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
		if !ok || key.NameExpr != nil || key.Name == name || overlaps(key.Name) || external(key.Name) || key.Name == "GLOBALS" || key.Name == "this" {
			return nil
		}
		keyName = key.Name
	}
	if pipeline {
		stmt, ok := tail[0].(*syntax.ExprStmt)
		if !ok {
			return nil
		}
		assign, ok := stmt.Expr.(*syntax.Assign)
		if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual {
			return nil
		}
		variable, ok := assign.Var.(*syntax.Variable)
		if !ok || variable.NameExpr != nil || variable.Name == name || overlaps(variable.Name) || variable.Name == keyName {
			return nil
		}
		if _, ok := tail[1].(*syntax.Foreach); !ok {
			return nil
		}
	}
	input := c.expression(loop.Expr)
	slot := c.nextSlot
	c.bind(name, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	for i, field := range fields {
		at := slot + 1
		if destructured {
			at = slot + 3 + i
		}
		c.bind(field.name, &pureExpr{kind: pureSlot, slot: at, cost: 1, depth: 1})
	}
	if keyName != "" {
		c.bind(keyName, &pureExpr{kind: pureSlot, slot: slot + 2, cost: 1, depth: 1})
	}
	var guard *pureExpr
	if skip != nil {
		guard = c.expression(skip)
		if guard == nil {
			return nil
		}
		guard = &pureExpr{kind: pureUnary, op: syntax.TExclaim, left: guard, projects: true, cost: guard.cost + 1, depth: guard.depth + 1}
	}
	if condition != nil {
		guard = c.expression(condition)
		if guard == nil {
			return nil
		}
	}
	var keyExpr *pureExpr
	if targetKey != nil {
		keyExpr = c.expression(targetKey)
		if keyExpr == nil {
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
		result.stopGuard = stop
		if destructured {
			result.destruct = fields
		}
		if build {
			result.kind = pureBuild
		}
		if guard != nil {
			if build {
				result.kind = pureFilterBuild
			} else {
				result.kind = pureGuardReduce
			}
			result.args = []*pureExpr{guard}
			result.cost += guard.cost
			result.depth = max(result.depth, guard.depth+1)
			if result.cost > pureLimit || result.depth > MaxDepth {
				return nil
			}
		}
		if keyExpr != nil {
			if len(result.args) == 0 {
				result.args = []*pureExpr{nil}
			}
			result.args = append(result.args, keyExpr)
			result.cost += keyExpr.cost
			result.depth = max(result.depth, keyExpr.depth+1)
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
	for _, field := range fields {
		final.bind(field.name, nil)
	}
	if keyName != "" {
		final.bind(keyName, nil)
	}
	final.bind(name, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	var returned *pureExpr
	if pipeline {
		final.nextSlot = slot + 3
		returned = final.statements(tail, MaxDepth)
	} else {
		returned = final.expression(ret.Expr)
	}
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
	if len(expr.destruct) != 0 {
		width = 3 + len(expr.destruct)
	}
	if expr.op == syntax.TDoubleArrow {
		width = max(width, 3)
	}
	r.copies += max(len(slots), expr.slot+width)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+width))
	copy(bound, slots)
	for _, key := range input.elements.order {
		item := input.elements.items[key]
		if _, scalar := knownScalar(item); !scalar && !((expr.kind == pureBuild || expr.kind == pureFilterBuild || len(expr.destruct) != 0) && pureReadOnlyInputs([]Value{item}, nil)) {
			return Value{}, false
		}
		bound[expr.slot], bound[expr.slot+1] = carry, item
		if len(expr.destruct) != 0 && !bindPureFields(expr.destruct, item, bound, expr.slot, r) {
			return Value{}, false
		}
		if expr.op == syntax.TDoubleArrow {
			keyValue := constantScalar{k: 's', s: key}
			if number, err := strconv.ParseInt(key, 10, 64); err == nil && strconv.FormatInt(number, 10) == key {
				keyValue = constantScalar{k: 'i', i: number}
			}
			bound[expr.slot+2] = scalarValue(keyValue)
		}

		if expr.kind == pureFilterBuild || expr.kind == pureGuardReduce {
			selected, valid := e.evalPure(expr.args[0], bound, r, depth-1)
			truth, known := pureTruthValue(selected)
			if !valid || !known {
				return Value{}, false
			}
			if !truth {
				if expr.stopGuard {
					break
				}
				continue
			}
		}
		if expr.kind == pureBuild || expr.kind == pureFilterBuild {
			key, keyed := "", len(expr.args) > 1
			if keyed {
				value, ok := e.evalPure(expr.args[1], bound, r, depth-1)
				var known bool
				key, known = pureRecordKey(value)
				if !ok || !known {
					return Value{}, false
				}
			}
			entry, valid := e.evalPure(expr.other, bound, r, depth-1)
			if !valid || !pureReadOnlyInputs([]Value{entry}, nil) {
				return Value{}, false
			}
			if keyed {
				if !putPureKey(&carry, key, entry, r) {
					return Value{}, false
				}
			} else if !appendPureEntry(&carry, entry, r) {
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
