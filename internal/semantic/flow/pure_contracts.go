package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

const (
	pureSlot uint8 = iota + 1
	pureLiteral
	pureUnary
	pureBinary
	pureCall
	pureLimit = 64
)

// Contracts contain values and instructions only. Costs count expanded call
// trees, so sharing a callee never bypasses the expression or recursion limits.
type pureExpr struct {
	kind          uint8
	weakArguments bool
	slot          int
	op            syntax.TokenKind
	literal       *constantScalar
	left, right   *pureExpr
	call          *pureContract
	args          []*pureExpr
	cost, depth   int
}

type pureParam struct {
	name               string
	typ                types.Type
	initial            *constantScalar
	optional, nullable bool
}

type pureContract struct {
	params     []pureParam
	expr       *pureExpr
	steps      []*pureExpr
	calls      []string
	result     types.Type
	weakReturn bool
	outer      int
}

type pureCompiler struct {
	env   *Env
	vars  map[string]*pureExpr
	work  int
	name  string
	calls []string
}

func pureConstant(value constantScalar) *pureExpr {
	literal := value
	return &pureExpr{kind: pureLiteral, literal: &literal, cost: 1, depth: 1}
}

func (e *Env) pureParameters(scope syntax.Node) ([]pureParam, bool) {
	source := syntax.VariableScopeParams(scope)
	if len(source) > 64 {
		return nil, false
	}
	lastRequired := -1
	for i, p := range source {
		if p.Default == nil {
			lastRequired = i
		}
	}
	params := make([]pureParam, len(source))
	for i, p := range source {
		if p.Var == nil || p.ByRef || p.Variadic {
			return nil, false
		}
		params[i] = pureParam{name: p.Var.Name, typ: types.FromNode(p.Type, nil), optional: p.Default != nil && i > lastRequired}
		if p.Default != nil {
			fuel := 256
			value, known := constantValue(p.Default, MaxDepth, &fuel)
			if !known {
				return nil, false
			}
			literal := value
			params[i].initial = &literal
			params[i].nullable = value.k == 'n'
		}
	}
	return params, true
}

func (e *Env) compilePure(scope syntax.Node, r *scopeResult, captures map[string]*pureExpr, outer int) *pureContract {
	if !r.complete || r.generator || byReference(scope) || r.argumentSensitive {
		return nil
	}
	params, known := e.pureParameters(scope)
	if !known {
		return nil
	}
	c := pureCompiler{env: e, vars: make(map[string]*pureExpr, len(captures)+len(params))}
	for name, value := range captures {
		c.vars[name] = value
	}
	for i, param := range params {
		c.vars[param.name] = &pureExpr{kind: pureSlot, slot: outer + i, cost: 1, depth: 1}
	}
	if e.summarizing != nil {
		c.name = e.summarizing.names[scope]
	}
	var result *pureExpr
	var steps []*pureExpr
	cost := 0
	switch body := syntax.VariableScopeBody(scope).(type) {
	case *syntax.Block:
		for i, stmt := range body.Stmts {
			switch n := stmt.(type) {
			case *syntax.Return:
				if i != len(body.Stmts)-1 || n.Expr == nil {
					return nil
				}
				result = c.expression(n.Expr)
			case *syntax.ExprStmt:
				assign, ok := n.Expr.(*syntax.Assign)
				if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual {
					return nil
				}
				variable, ok := assign.Var.(*syntax.Variable)
				if !ok || variable.NameExpr != nil || external(variable.Name) || variable.Name == "GLOBALS" || variable.Name == "this" {
					return nil
				}
				value := c.expression(assign.Value)
				if value == nil {
					return nil
				}
				cost += value.cost
				steps = append(steps, value)
				c.vars[variable.Name] = value
			default:
				return nil
			}
		}
	default:
		if _, arrow := scope.(*syntax.ArrowFunction); !arrow {
			return nil
		}
		result = c.expression(body.(syntax.Expr))
	}
	if result == nil || cost+result.cost > pureLimit {
		return nil
	}
	return &pureContract{params: params, expr: result, steps: steps, calls: c.calls, result: types.FromNode(returnDeclaration(scope), nil), weakReturn: !syntax.StrictTypes(e.file), outer: outer}
}

func (c *pureCompiler) expression(expr syntax.Expr) *pureExpr {
	c.work++
	if c.work > pureLimit {
		return nil
	}
	expr = syntax.UnwrapParens(expr)
	switch n := expr.(type) {
	case *syntax.Variable:
		if n.NameExpr == nil {
			return c.vars[n.Name]
		}
	case *syntax.Literal, *syntax.ConstFetch:
		fuel := 256
		if value, known := constantValue(expr, MaxDepth, &fuel); known {
			return pureConstant(value)
		}
	case *syntax.Unary:
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TPlus, syntax.TMinus:
			child := c.expression(n.Expr)
			if child != nil && child.cost < pureLimit && child.depth < MaxDepth {
				return &pureExpr{kind: pureUnary, op: n.Op.Kind, left: child, cost: child.cost + 1, depth: child.depth + 1}
			}
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDot, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
			left, right := c.expression(n.Left), c.expression(n.Right)
			if left != nil && right != nil && left.cost+right.cost < pureLimit && max(left.depth, right.depth) < MaxDepth {
				return &pureExpr{kind: pureBinary, op: n.Op.Kind, left: left, right: right, cost: left.cost + right.cost + 1, depth: max(left.depth, right.depth) + 1}
			}
		}
	case *syntax.FuncCall:
		if _, named := n.Name.(*syntax.Name); named {
			return c.call(n)
		}
	}
	return nil
}

func (c *pureCompiler) call(node *syntax.FuncCall) *pureExpr {
	_, name, known := c.env.declaration(node)
	if !known {
		return nil
	}
	summary, known := c.env.contract(name)
	if !known || !summary.Fixed || summary.Generator || summary.pure == nil {
		return nil
	}
	callee := summary.pure
	key := strings.ToLower(name)
	if c.name != "" && (key == c.name || slices.Contains(callee.calls, c.name)) {
		return nil
	}
	c.calls = append(c.calls, key)
	c.calls = append(c.calls, callee.calls...)
	args := make([]*pureExpr, len(callee.params))
	position := 0
	named := false
	if node.Args != nil {
		for _, expr := range node.Args.Args {
			arg, ok := expr.(*syntax.Arg)
			if !ok || arg.ByRef || arg.Unpack {
				return nil
			}
			index := position
			if arg.Name != nil {
				if c.env.types.PHP.Below(phpversion.PHP80) {
					return nil
				}
				named, index = true, -1
				for i, p := range callee.params {
					if p.name == arg.Name.Value {
						index = i
						break
					}
				}
			} else {
				if named {
					return nil
				}
				position++
			}
			if index < 0 || index >= len(args) || args[index] != nil {
				return nil
			}
			args[index] = c.expression(arg.Value)
			if args[index] == nil {
				return nil
			}
		}
	}
	cost, depth := callee.expr.cost+1, callee.expr.depth+1
	for _, step := range callee.steps {
		cost += step.cost
		depth = max(depth, step.depth+1)
	}
	for i, param := range callee.params {
		if args[i] == nil {
			if !param.optional || param.initial == nil {
				return nil
			}
			args[i] = pureConstant(*param.initial)
		}
		cost += args[i].cost
		depth = max(depth, args[i].depth+1)
	}
	if cost > pureLimit || depth > MaxDepth {
		return nil
	}
	return &pureExpr{kind: pureCall, call: callee, args: args, cost: cost, depth: depth, weakArguments: !syntax.StrictTypes(c.env.file)}
}
