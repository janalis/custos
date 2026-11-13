package flow

import (
	"maps"
	"math"
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
	case *syntax.Switch:
		if len(stmts) != 1 {
			return nil
		}
		return c.integerSwitch(n)
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
		switch assign.Var.(type) {
		case *syntax.List, *syntax.Array:
			return c.destructure(assign, stmts[1:], depth)
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
