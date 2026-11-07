package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// Quiet reads are separate instructions: an ordinary missing offset can warn.
// Clone the path rather than changing a binding shared with an ordinary read.
func quietPure(expr *pureExpr) *pureExpr {
	if expr == nil || expr.kind != pureIndex {
		return expr
	}
	copy := *expr
	copy.kind = pureQuietIndex
	copy.left = quietPure(expr.left)
	return &copy
}

func (c *pureCompiler) isset(node *syntax.Isset) *pureExpr {
	var result *pureExpr
	for _, variable := range node.Vars {
		value := quietPure(c.expression(variable))
		test := pureNode(pureNonNull, value, pureConstant(constantScalar{k: 'n'}), nil)
		if test == nil {
			return nil
		}
		if result == nil {
			result = test
		} else {
			result = pureNode(pureShort, result, test, nil)
			if result == nil {
				return nil
			}
			result.op = syntax.TBooleanAnd
		}
	}
	return result
}

func (c *pureCompiler) keyExists(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 2 {
		return nil
	}
	roles := []pureParam{{name: params[0].Name}, {name: params[1].Name}}
	args := c.arguments(node, roles, params)
	if args == nil {
		return nil
	}
	return pureNode(pureExists, args[1], args[0], nil)
}

func (c *pureCompiler) staticCall(node *syntax.StaticCall) *pureExpr {
	class, named := node.Class.(*syntax.Name)
	method, fixed := node.Name.(*syntax.Identifier)
	if !named || !fixed || strings.EqualFold(class.Value, "static") {
		return nil
	}
	target, known := c.env.closedMethod(node, c.env.named(class), method.Value, true)
	if !known || target.summary.pure == nil {
		return nil
	}
	return c.projectCall(node, target.method.Class+"::"+target.method.Name, target.summary.pure, target.method.Params)
}

func (c *pureCompiler) match(node *syntax.Match) *pureExpr {
	if c.env.types.PHP.Below(phpversion.PHP80) {
		return nil
	}
	subject := c.expression(node.Cond)
	if subject == nil {
		return nil
	}
	result := &pureExpr{kind: pureMatch, left: subject, cost: subject.cost + 1, depth: subject.depth + 1, projects: true}
	for _, arm := range node.Arms {
		body := c.expression(arm.Body)
		if body == nil {
			return nil
		}
		if len(arm.Conds) == 0 {
			if result.other != nil {
				return nil
			}
			result.other = body
			result.cost += body.cost
			result.depth = max(result.depth, body.depth+1)
			result.projects = result.projects || body.projects
		}
		for _, cond := range arm.Conds {
			test := c.expression(cond)
			if test == nil {
				return nil
			}
			result.args = append(result.args, test, body)
			result.cost += test.cost + body.cost
			result.depth = max(result.depth, max(test.depth, body.depth)+1)
			result.projects = result.projects || test.projects || body.projects
		}
	}
	if result.cost > pureLimit || result.depth > MaxDepth {
		return nil
	}
	return result
}

func (e *Env) evalPureMatch(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	value, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	for i := 0; i < len(expr.args); i += 2 {
		test, valid := e.evalPure(expr.args[i], slots, r, depth-1)
		if !valid {
			return Value{}, false
		}
		equal, known := pureQueryIdentity(value, test, r, depth-1)
		if !known {
			return Value{}, false
		}
		if equal {
			return e.evalPure(expr.args[i+1], slots, r, depth-1)
		}
	}
	if expr.other == nil {
		return Value{}, false
	}
	return e.evalPure(expr.other, slots, r, depth-1)
}

func (c *pureCompiler) instanceCall(node *syntax.MethodCall) *pureExpr {
	id, fixed := node.Name.(*syntax.Identifier)
	if !fixed || node.NullSafe || plainVariable(node.Var) != "this" {
		return nil
	}
	scope := syntax.EnclosingVariableScope(node)
	method, ok := scope.(*syntax.Method)
	if !ok || method.Modifiers.Has(syntax.TStatic) {
		return nil
	}
	// In a non-static method, the lexical class supplies the native this
	// upper bound. Closed dispatch still requires an immutable method target.
	class := c.env.types.ClassFQN(syntax.EnclosingClass(node))
	target, known := c.env.closedMethod(node, class, id.Value, false)
	if !known || target.method.Static || target.method.ByRef || target.summary.pure == nil {
		return nil
	}
	return c.projectCall(node, target.method.Class+"::"+target.method.Name, target.summary.pure, target.method.Params)
}
