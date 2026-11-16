package flow

import (
	"custos/internal/php/syntax"
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
	pureJoin
	pureSplit
	pureColumn
	pureMerge
	pureInteger
	pureReplace
	pureRepeat
	pureReverse
	pureSwitch
	pureCombine
	pureArrayReplace
	pureDestructure
	pureFill
	pureFillKeys
	pureFlip
	pureChunk
	pureSum
	pureCase
	pureTrimChars
	pureStrPad
	pureArrayPad
	pureRange
	pureUnique
	pureSearch
	pureExtreme
	pureProduct
	pureTextScan
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
	prepared      bool
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
