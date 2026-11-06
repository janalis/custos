package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// Compile only a direct noncapturing arrow; never publish a callable value.
func (c *pureCompiler) inlineArrow(node *syntax.ArrowFunction) *pureExpr {
	if c.inlineDepth >= MaxDepth || c.env.types.PHP.Below(phpversion.PHP74) || node.ByRef || len(node.Attrs) != 0 {
		return nil
	}
	seen := map[string]bool{}
	for _, param := range node.Params {
		if param.Var == nil || param.Var.Name == "" || param.Var.NameExpr != nil || external(param.Var.Name) || param.Var.Name == "this" || param.Var.Name == "GLOBALS" || seen[param.Var.Name] {
			return nil
		}
		seen[param.Var.Name] = true
	}
	params, ok := c.env.pureParameters(node)
	if !ok {
		return nil
	}
	child := pureCompiler{env: c.env, name: c.name, inlineDepth: c.inlineDepth + 1, nextSlot: len(params)}
	for i, param := range params {
		child.bind(param.name, &pureSlots[i])
	}
	body := child.expression(node.Expr)
	c.work += child.work
	if body == nil || body.cost > pureLimit || body.depth > MaxDepth || c.work > pureLimit {
		return nil
	}
	contract := &pureContract{params: params, expr: body, calls: child.calls, result: types.FromNode(node.ReturnType, nil), weakReturn: !syntax.StrictTypes(c.env.file)}
	return &pureExpr{kind: pureInline, call: contract, cost: 1, depth: 1}
}

func (c *pureCompiler) fold(node *syntax.FuncCall, params []index.Param) *pureExpr {
	if len(params) != 3 {
		return nil
	}
	roles := []pureParam{{name: params[0].Name}, {name: params[1].Name}, {name: params[2].Name}}
	args := c.arguments(node, roles, params)
	if args == nil || args[0] == nil || args[1] == nil || args[2] == nil || args[1].kind != pureLiteral || args[1].literal.k != 's' {
		return nil
	}
	key := callbackName(args[1].literal.s)
	invoke := c.namedCallback(key, 2)
	if invoke == nil {
		return nil
	}
	result := pureNode(pureFold, args[0], args[2], invoke)
	if result != nil {
		result.projects = true
	}
	return result
}

func (e *Env) evalPureFold(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	input, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok || !purePlainRecord(input) || len(input.elements.order) != len(input.elements.items) || len(input.elements.order) > pureLimit {
		return Value{}, false
	}
	carry, ok := e.evalPure(expr.right, slots, r, depth-1)
	if _, known := knownScalar(carry); !ok || !known || !pureReadOnlyInputs([]Value{carry}, nil) {
		return Value{}, false
	}
	for _, key := range input.elements.order {
		carry, ok = e.evalPure(expr.other, []Value{carry, input.elements.items[key]}, r, depth-1)
		if _, known := knownScalar(carry); !ok || !known || !pureReadOnlyInputs([]Value{carry}, nil) {
			return Value{}, false
		}
	}
	return carry, true
}
