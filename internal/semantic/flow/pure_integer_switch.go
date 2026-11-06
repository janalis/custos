package flow

import "custos/internal/php/syntax"

func (c *pureCompiler) integerSwitch(node *syntax.Switch) *pureExpr {
	subject := c.expression(node.Cond)
	if subject == nil || len(node.Cases) > pureLimit {
		return nil
	}
	result := &pureExpr{kind: pureSwitch, left: subject, projects: true, cost: subject.cost + 1, depth: subject.depth + 1}
	for _, arm := range node.Cases {
		if arm == nil || len(arm.Stmts) != 1 {
			return nil
		}
		returned, ok := singlePureStatement(arm.Stmts[0]).(*syntax.Return)
		if !ok {
			return nil
		}
		value := pureConstant(constantScalar{k: 'n'})
		if returned.Expr != nil {
			value = c.expression(returned.Expr)
		}
		if value == nil {
			return nil
		}
		result.cost += value.cost
		result.depth = max(result.depth, value.depth+1)
		if arm.Cond == nil {
			if result.other != nil {
				return nil
			}
			result.other = value
		} else {
			test := c.expression(arm.Cond)
			if test == nil {
				return nil
			}
			result.args = append(result.args, test, value)
			result.cost += test.cost
			result.depth = max(result.depth, test.depth+1)
		}
	}
	if result.other == nil || result.cost > pureLimit || result.depth > MaxDepth {
		return nil
	}
	return result
}

func (e *Env) evalIntegerSwitch(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	subject, ok := e.evalPure(expr.left, slots, r, depth-1)
	value, known := knownScalar(subject)
	if !ok || !known || value.k != 'i' {
		return Value{}, false
	}
	for i := 0; i < len(expr.args); i += 2 {
		tested, ok := e.evalPure(expr.args[i], slots, r, depth-1)
		test, known := knownScalar(tested)
		if !ok || !known || test.k != 'i' {
			return Value{}, false
		}
		if value.i == test.i {
			return e.evalPure(expr.args[i+1], slots, r, depth-1)
		}
	}
	return e.evalPure(expr.other, slots, r, depth-1)
}
