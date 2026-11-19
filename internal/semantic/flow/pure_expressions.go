package flow

import (
	"slices"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
)

func (c *pureCompiler) expression(expr syntax.Expr) *pureExpr {
	c.work++
	if c.work > pureLimit {
		return nil
	}
	expr = syntax.UnwrapParens(expr)
	if c.defaults {
		fuel := 256
		if scalar, ok := constantValue(expr, MaxDepth, &fuel); ok {
			return pureConstant(scalar)
		}
		if _, ok := expr.(*syntax.Array); !ok {
			return nil
		}
	}
	switch n := expr.(type) {
	case *syntax.Closure:
		if c.inlineCallbacks {
			return c.inlineClosure(n)
		}
	case *syntax.ArrowFunction:
		if c.inlineCallbacks {
			return c.inlineArrow(n)
		}
	case *syntax.Match:
		return c.match(n)
	case *syntax.Empty:
		return c.empty(n)
	case *syntax.Isset:
		return c.isset(n)
	case *syntax.MethodCall:
		return c.instanceCall(n)
	case *syntax.StaticCall:
		return c.staticCall(n)
	case *syntax.Ternary:
		condition := c.expression(n.Cond)
		yes := condition
		if n.Then != nil {
			yes = c.expression(n.Then)
		}
		return pureNode(pureChoice, condition, yes, c.expression(n.Else))
	case *syntax.Array:
		return c.array(n)
	case *syntax.ArrayDimFetch:
		if n.Dim != nil && !n.Curly {
			return pureNode(pureIndex, c.expression(n.Var), c.expression(n.Dim), nil)
		}
	case *syntax.Variable:
		if n.NameExpr == nil {
			for _, binding := range c.vars {
				if binding.name == n.Name {
					return binding.value
				}
			}
			return c.extra[n.Name]
		}
	case *syntax.ConstFetch:
		fuel := 256
		if value, known := constantValue(expr, MaxDepth, &fuel); known {
			return pureConstant(value)
		}
		return c.nativeConstant(n)
	case *syntax.Literal:
		fuel := 256
		if value, known := constantValue(expr, MaxDepth, &fuel); known {
			return pureConstant(value)
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TTilde {
			return c.integer(n.Op.Kind, n.Expr, nil)
		}
		if n.Op.Kind == syntax.TBoolCast || n.Op.Kind == syntax.TIntCast || n.Op.Kind == syntax.TStringCast {
			return c.cast(n)
		}
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TPlus, syntax.TMinus:
			child := c.expression(n.Expr)
			if child != nil && child.cost < pureLimit && child.depth < MaxDepth {
				return &pureExpr{kind: pureUnary, op: n.Op.Kind, left: child, projects: n.Op.Kind == syntax.TExclaim || child.projects, cost: child.cost + 1, depth: child.depth + 1}
			}
		}
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TDiv, syntax.TMod, syntax.TAmpersand, syntax.TBar, syntax.TCaret, syntax.TSl, syntax.TSr:
			return c.integer(n.Op.Kind, n.Left, n.Right)
		}
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr, syntax.TCoalesce:
			if n.Op.Kind == syntax.TCoalesce && c.env.types.PHP.Below(phpversion.PHP70) {
				return nil
			}
			left := c.expression(n.Left)
			if n.Op.Kind == syntax.TCoalesce {
				left = quietPure(left)
			}
			expr := pureNode(pureShort, left, c.expression(n.Right), nil)
			if expr != nil {
				expr.op = n.Op.Kind
			}
			return expr
		}
		switch n.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDot, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
			left, right := c.expression(n.Left), c.expression(n.Right)
			if left != nil && right != nil && left.cost+right.cost < pureLimit && max(left.depth, right.depth) < MaxDepth {
				return &pureExpr{kind: pureBinary, op: n.Op.Kind, left: left, right: right, projects: left.projects || right.projects, cost: left.cost + right.cost + 1, depth: max(left.depth, right.depth) + 1}
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
	params, name, known := c.env.declaration(node)
	if native, builtin := c.env.callName(node); builtin {
		return c.native(node, native, params)
	}
	if !known {
		return nil
	}
	summary, known := c.env.contract(name)
	if !known || !summary.Fixed || summary.Generator || summary.pure == nil {
		return nil
	}
	return c.projectCall(node, name, summary.pure, params)
}

func (c *pureCompiler) projectCall(node syntax.Node, name string, callee *pureContract, params []index.Param) *pureExpr {
	key := strings.ToLower(name)
	if c.name != "" && (key == c.name || slices.Contains(callee.calls, c.name)) {
		return nil
	}
	c.calls = append(c.calls, key)
	c.calls = append(c.calls, callee.calls...)
	args := c.arguments(node, callee.params, params)
	if args == nil {
		return nil
	}
	cost, depth := callee.expr.cost+1, callee.expr.depth+1
	for i, param := range callee.params {
		if args[i] == nil {
			if !param.optional || param.initial == nil {
				return nil
			}
			args[i] = param.initial
		}
		cost += args[i].cost
		depth = max(depth, args[i].depth+1)
	}
	if cost > pureLimit || depth > MaxDepth {
		return nil
	}
	return &pureExpr{kind: pureCall, call: callee, args: args, projects: callee.expr.projects, cost: cost, depth: depth, weakArguments: !syntax.StrictTypes(c.env.file)}
}
