package flow

import "custos/internal/php/syntax"

// Recognize only an initialized carry, a single assignment, and its final
// return. No loop-local state or control flow escapes this contract.
func (c *pureCompiler) reduction(name string, initial *pureExpr, loop *syntax.Foreach, tail []syntax.Stmt) *pureExpr {
	if loop.ByRef || loop.Key != nil || len(tail) != 1 {
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
	if !ok || update.ByRef || update.Op.Kind != syntax.TEqual {
		return nil
	}
	target, ok := update.Var.(*syntax.Variable)
	if !ok || target.NameExpr != nil || target.Name != name {
		return nil
	}
	input := c.expression(loop.Expr)
	slot := c.nextSlot
	c.bind(name, &pureExpr{kind: pureSlot, slot: slot, cost: 1, depth: 1})
	c.bind(item.Name, &pureExpr{kind: pureSlot, slot: slot + 1, cost: 1, depth: 1})
	result := pureNode(pureReduce, input, initial, c.expression(update.Value))
	if result != nil {
		result.slot = slot
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
	if _, scalar := knownScalar(carry); !ok || !scalar {
		return Value{}, false
	}
	r.copies += max(len(slots), expr.slot+2)
	if r.copies > MaxCopied {
		return Value{}, false
	}
	bound := make([]Value, max(len(slots), expr.slot+2))
	copy(bound, slots)
	for _, key := range input.elements.order {
		item := input.elements.items[key]
		if _, scalar := knownScalar(item); !scalar {
			return Value{}, false
		}
		bound[expr.slot], bound[expr.slot+1] = carry, item
		carry, ok = e.evalPure(expr.other, bound, r, depth-1)
		if _, scalar := knownScalar(carry); !ok || !scalar {
			return Value{}, false
		}
	}
	return carry, true
}
