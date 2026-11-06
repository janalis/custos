package flow

import "custos/internal/php/syntax"

func singlePureStatement(stmt syntax.Stmt) syntax.Stmt {
	if block, ok := stmt.(*syntax.Block); ok {
		if len(block.Stmts) != 1 {
			return nil
		}
		return block.Stmts[0]
	}
	return stmt
}

func (c *pureCompiler) find(loop *syntax.Foreach, tail []syntax.Stmt) *pureExpr {
	if loop.ByRef || len(tail) != 1 {
		return nil
	}
	fallback, ok := tail[0].(*syntax.Return)
	if !ok {
		return nil
	}
	branch, ok := singlePureStatement(loop.Body).(*syntax.If)
	if !ok || branch.Else != nil || len(branch.ElseIfs) != 0 {
		return nil
	}
	returned, ok := singlePureStatement(branch.Body).(*syntax.Return)
	if !ok {
		return nil
	}
	item := plainVariable(loop.Value)
	key := ""
	if item == "" || external(item) || item == "GLOBALS" || item == "this" {
		return nil
	}
	if loop.Key != nil {
		key = plainVariable(loop.Key)
		if key == "" || key == item || external(key) || key == "GLOBALS" || key == "this" {
			return nil
		}
	}
	input := c.expression(loop.Expr)
	// Remove loop bindings before compiling fallback: empty inputs have none.
	originalItem, originalKey := c.binding(item), c.binding(key)
	c.bind(item, nil)
	if key != "" {
		c.bind(key, nil)
	}
	otherwise := c.expression(fallback.Expr)
	c.bind(item, originalItem)
	if key != "" {
		c.bind(key, originalKey)
	}
	slot := c.nextSlot
	c.bind(item, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	if key != "" {
		c.bind(key, &pureExpr{kind: pureSlot, slot: slot + 1, cost: 1, depth: 1})
	}
	condition := c.expression(branch.Cond)
	value := c.expression(returned.Expr)
	result := pureNode(pureFind, input, condition, value)
	if result == nil || value == nil || otherwise == nil {
		return nil
	}
	result.args = []*pureExpr{otherwise}
	result.cost += otherwise.cost
	result.depth = max(result.depth, otherwise.depth+1)
	if result.cost > pureLimit || result.depth > MaxDepth {
		return nil
	}
	result.slot = slot
	result.projects = true
	if key != "" {
		result.op = syntax.TDoubleArrow
	}
	return result
}

func (c *pureCompiler) binding(name string) *pureExpr {
	for _, entry := range c.vars {
		if entry.name == name {
			return entry.value
		}
	}
	return c.extra[name]
}

func (e *Env) evalPureFind(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	input, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok || !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	r.copies += max(len(slots), expr.slot+2)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+2))
	copy(bound, slots)
	for _, key := range input.elements.order {
		bound[expr.slot] = input.elements.items[key]
		if expr.op == syntax.TDoubleArrow {
			bound[expr.slot+1] = scalarValue(pureKeyScalar(key))
		}
		value, ok := e.evalPure(expr.right, bound, r, depth-1)
		selected, known := pureTruthValue(value)
		if !ok || !known {
			return Value{}, false
		}
		if selected {
			return e.evalPure(expr.other, bound, r, depth-1)
		}
	}
	return e.evalPure(expr.args[0], slots, r, depth-1)
}
