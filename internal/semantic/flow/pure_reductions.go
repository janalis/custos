package flow

import (
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
	returned, ok := ret.Expr.(*syntax.Variable)
	if !ok || returned.NameExpr != nil || returned.Name != name {
		return nil
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
		if keyName != "" {
			result.op = syntax.TDoubleArrow
		}
		result.projects = true
	}
	return result
}

func (e *Env) evalPureReduction(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	input, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok || !input.Complete || input.replaced || input.Invalidated != 0 || !input.plain || !input.deep || input.elements == nil || !input.elements.whole || input.elements.optional || input.elements.hasResources || len(input.elements.order) != len(input.elements.items) {
		return Value{}, false
	}
	carry, ok := e.evalPure(expr.right, slots, r, depth-1)
	if _, scalar := knownScalar(carry); !ok || !scalar && !(expr.kind == pureBuild && purePlainRecord(carry)) {
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
		if _, scalar := knownScalar(item); !scalar && !(expr.kind == pureBuild && pureReadOnlyInputs([]Value{item}, nil)) {
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

		if expr.kind == pureBuild {
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
