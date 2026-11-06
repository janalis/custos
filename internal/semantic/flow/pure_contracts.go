package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

const (
	pureSlot uint8 = iota + 1
	pureLiteral
	pureUnary
	pureBinary
	pureCall
	pureChoice
	pureSequence
	pureShort
	pureArray
	pureIndex
	pureLength
	pureTrim
	pureMatch
	pureQuietIndex
	pureExists
	pureReduce
	pureNonNull
	pureCount
	pureMember
	pureStringPredicate
	pureSpreadArray
	pureBuild
	pureFilterBuild
	pureLet
	pureOrderedArray
	pureTypeTest
	pureSubstring
	pureEmpty
	pureCast
	pureHigher
	pureFind
	pureGuardReduce
	pureInline
	pureFold
	pureSlice
	pureFilteredKeys
	pureLimit = 64
)

// Contracts contain values and instructions only. Costs count expanded call
// trees, so sharing a callee never bypasses the expression or recursion limits.
type pureExpr struct {
	kind          uint8
	weakArguments bool
	projects      bool
	appendKey     bool
	stopGuard     bool
	destruct      []pureField
	slot          int
	op            syntax.TokenKind
	literal       *constantScalar
	left, right   *pureExpr
	other         *pureExpr
	call          *pureContract
	args          []*pureExpr
	cost, depth   int
}

type pureParam struct {
	name               string
	typ                types.Type
	initial            *pureExpr
	optional, nullable bool
}

type pureContract struct {
	params     []pureParam
	expr       *pureExpr
	calls      []string
	result     types.Type
	weakReturn bool
	outer      int
}

// Slot instructions contain no declaration or invocation state and can be
// shared across contracts. Outer and invocation parameters each have 64 slots.
var pureSlots = func() [128]pureExpr {
	var slots [128]pureExpr
	for i := range slots {
		slots[i] = pureExpr{kind: pureSlot, slot: i, cost: 1, depth: 1}
	}
	return slots
}()

type pureBinding struct {
	name  string
	value *pureExpr
}

type pureCompiler struct {
	env             *Env
	vars            []pureBinding
	extra           map[string]*pureExpr
	work            int
	name            string
	calls           []string
	defaults        bool
	nextSlot        int
	inlineCallbacks bool
	inlineDepth     int
}

// Small scopes keep bindings on the compiler's stack. Larger scopes retain
// constant-time overflow lookup without allocating a map for every function.
func (c *pureCompiler) bind(name string, value *pureExpr) {
	for i := range c.vars {
		if c.vars[i].name == name {
			c.vars[i].value = value
			return
		}
	}
	if len(c.vars) < 8 {
		c.vars = append(c.vars, pureBinding{name: name, value: value})
		return
	}
	if c.extra == nil {
		c.extra = make(map[string]*pureExpr)
	}
	c.extra[name] = value
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
			c := pureCompiler{env: e, defaults: true}
			value := c.expression(p.Default)
			if value == nil {
				return nil, false
			}
			params[i].initial = value
			params[i].nullable = value.kind == pureLiteral && value.literal.k == 'n'
		}
	}
	return params, true
}

func (e *Env) compilePure(scope syntax.Node, r *scopeResult, captures map[string]*pureExpr, outer int) *pureContract {
	if !r.complete || r.generator || byReference(scope) || r.argumentSensitive {
		return nil
	}
	source := syntax.VariableScopeParams(scope)
	if len(source) > 64 {
		return nil
	}
	var locals [8]pureBinding
	c := pureCompiler{env: e, vars: locals[:0], nextSlot: outer + len(source)}
	for name, value := range captures {
		c.bind(name, value)
	}
	for i, param := range source {
		if param.Var == nil || param.ByRef || param.Variadic {
			return nil
		}
		c.bind(param.Var.Name, &pureSlots[outer+i])
	}
	if e.summarizing != nil {
		c.name = e.summarizing.names[scope]
	}
	var result *pureExpr
	switch body := syntax.VariableScopeBody(scope).(type) {
	case *syntax.Block:
		result = c.statements(body.Stmts, MaxDepth)
	default:
		if _, arrow := scope.(*syntax.ArrowFunction); !arrow {
			return nil
		}
		result = c.expression(body.(syntax.Expr))
	}
	if result == nil || result.cost > pureLimit || result.depth > MaxDepth || c.work > pureLimit {
		return nil
	}
	// Parameter types and default programs are needed only for a supported body.
	// Most declarations cannot form a pure contract. Avoid building unused defaults.
	params, known := e.pureParameters(scope)
	if !known {
		return nil
	}
	return &pureContract{params: params, expr: result, calls: c.calls, result: types.FromNode(returnDeclaration(scope), nil), weakReturn: !syntax.StrictTypes(e.file), outer: outer}
}

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
