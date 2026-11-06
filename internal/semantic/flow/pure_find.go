package flow

import (
	"slices"

	"custos/internal/php/syntax"
)

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
	fields, destructured := c.loopFields(loop.Value, "")
	if fields == nil {
		return nil
	}
	key := ""
	if loop.Key != nil {
		key = plainVariable(loop.Key)
		if key == "" || external(key) || key == "GLOBALS" || key == "this" || slices.ContainsFunc(fields, func(field pureField) bool { return field.name == key }) {
			return nil
		}
	}
	input := c.expression(loop.Expr)
	// Remove loop bindings before compiling fallback: empty inputs have none.
	originals := make([]*pureExpr, len(fields))
	for i, field := range fields {
		originals[i] = c.binding(field.name)
		c.bind(field.name, nil)
	}
	originalKey := c.binding(key)
	if key != "" {
		c.bind(key, nil)
	}
	otherwise := c.expression(fallback.Expr)
	for i, field := range fields {
		c.bind(field.name, originals[i])
	}
	if key != "" {
		c.bind(key, originalKey)
	}
	slot := c.nextSlot
	for i, field := range fields {
		at := slot
		if destructured {
			at = slot + 3 + i
		}
		c.bind(field.name, &pureExpr{kind: pureSlot, slot: at, cost: 1, depth: 1})
	}
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
	if destructured {
		result.destruct = fields
	}
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
	width := 2
	if len(expr.destruct) != 0 {
		width = 3 + len(expr.destruct)
	}
	r.copies += max(len(slots), expr.slot+width)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+width))
	copy(bound, slots)
	for _, key := range input.elements.order {
		bound[expr.slot] = input.elements.items[key]
		if len(expr.destruct) != 0 && !bindPureFields(expr.destruct, input.elements.items[key], bound, expr.slot, r) {
			return Value{}, false
		}
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
