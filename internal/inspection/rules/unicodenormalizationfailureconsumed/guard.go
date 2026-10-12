package unicodenormalizationfailureconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard ties strict failure checks to one consumed value: input is the read
// operand and origin the normalization call that reaches it on every path.
type guard struct {
	ctx    *analysis.Context
	input  syntax.Expr
	origin syntax.Expr
}

// dominated proves that a strict false check or an is_string test keeps the
// failure value away from the consumption. Truthiness, loose comparison and
// assert() stay unknown rather than guessed safe.
func (g guard) dominated() bool {
	var child syntax.Node = g.input
	for p := g.input.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Ternary:
			if (x.Then != nil && child == syntax.Node(x.Then) && g.holds(x.Cond, true)) || (child == syntax.Node(x.Else) && g.holds(x.Cond, false)) {
				return true
			}
		case *syntax.Binary:
			if truth, ok := shortCircuit(x); ok && child == syntax.Node(x.Right) && g.holds(x.Left, truth) {
				return true
			}
		case *syntax.MatchArm:
			if g.matched(x, child) {
				return true
			}
		case *syntax.If:
			if g.branched(x, child) {
				return true
			}
		case *syntax.ElseIf:
			if child == syntax.Node(x.Body) && g.holds(x.Cond, true) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && g.exited(st) {
			return true
		}
	}
	return false
}

// shortCircuit returns the truth the left operand must have for the right
// operand to be evaluated.
func shortCircuit(b *syntax.Binary) (truth, ok bool) {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return true, true
	case syntax.TBooleanOr, syntax.TOr:
		return false, true
	}
	return false, false
}

// branched handles the branches of an if statement: its own body requires the
// condition, a later branch requires every earlier condition to have failed.
func (g guard) branched(x *syntax.If, child syntax.Node) bool {
	if child == syntax.Node(x.Body) {
		return g.holds(x.Cond, true)
	}
	if child == syntax.Node(x.Cond) {
		return false
	}
	if g.holds(x.Cond, false) {
		return true
	}
	for _, alternative := range x.ElseIfs {
		if child == syntax.Node(alternative) {
			break
		}
		if g.holds(alternative.Cond, false) {
			return true
		}
	}
	return false
}

// matched handles match arms. An arm of match (true) is entered only when its
// condition is identical to true and every earlier condition was not; an arm
// after a literal false condition cannot see a false subject.
func (g guard) matched(arm *syntax.MatchArm, child syntax.Node) bool {
	m := arm.Parent().(*syntax.Match)
	subject, boolean := astquery.BoolConst(syntax.UnwrapParens(m.Cond))
	switch {
	case boolean && !subject:
		return false
	case !boolean && !g.same(m.Cond):
		return false
	}
	if boolean && child == syntax.Node(arm.Body) && len(arm.Conds) == 1 && g.holds(arm.Conds[0], true) {
		return true
	}
	for _, earlier := range m.Arms {
		if earlier == arm {
			break
		}
		for _, cond := range earlier.Conds {
			if boolean && g.holds(cond, false) {
				return true
			}
			if value, literal := astquery.BoolConst(syntax.UnwrapParens(cond)); !boolean && literal && !value {
				return true
			}
		}
	}
	return false
}

// exited looks for an earlier sibling of st that leaves the path whenever the
// value is false: a terminating if without alternatives. A short-circuit
// statement ending in exit or throw is a fact of the shared flow analysis,
// asked before this guard.
func (g guard) exited(st syntax.Stmt) bool {
	list, index, ok := astquery.StmtList(g.ctx.File, st)
	if !ok {
		return false
	}
	for _, previous := range list[:index] {
		if x, ok := previous.(*syntax.If); ok && x.Else == nil && len(x.ElseIfs) == 0 && syntax.Terminates(x.Body) && g.holds(x.Cond, false) {
			return true
		}
	}
	return false
}

// holds proves that cond evaluating to truth implies the value is not false.
// Every accepted condition is boolean-typed, so "not identical to true" and
// "false" coincide for match (true) arms.
func (g guard) holds(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.holds(x.Expr, !truth)
	case *syntax.FuncCall:
		return truth && g.stringTest(x)
	case *syntax.Binary:
		if required, ok := shortCircuit(x); ok {
			return required == truth && (g.holds(x.Left, truth) || g.holds(x.Right, truth))
		}
		if x.Op.Kind != syntax.TIsIdentical && x.Op.Kind != syntax.TIsNotIdentical {
			return false
		}
		equal := (x.Op.Kind == syntax.TIsIdentical) == truth
		for _, pair := range [][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
			literal, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1]))
			if !ok {
				continue
			}
			if !literal && !equal && g.same(pair[0]) {
				return true
			}
			if call, ok := syntax.UnwrapParens(pair[0]).(*syntax.FuncCall); ok && literal == equal && g.stringTest(call) {
				return true
			}
		}
	}
	return false
}

func (g guard) stringTest(call *syntax.FuncCall) bool {
	return semanticquery.NativeBuiltin(g.ctx, call, "is_string") && g.same(semanticquery.CallArgument(call.Args, 0, "value"))
}

// same proves that a checked operand is the consumed value: the same variable
// (or a plain assignment to it) carrying the same reaching normalization call,
// checked after that call ran. A repeated call or a rebinding is a new value;
// compound and by-reference assignments leave no agreed reaching value.
func (g guard) same(operand syntax.Expr) bool {
	if operand == nil {
		return false
	}
	operand = syntax.UnwrapParens(operand)
	target := operand
	if a, ok := operand.(*syntax.Assign); ok {
		target = a.Var
	}
	if !astquery.Equivalent(g.ctx.File, g.input, target) || g.origin.Span().End > operand.Span().End {
		return false
	}
	checked, consumed := g.ctx.Flow().Value(operand), g.ctx.Flow().Value(g.input)
	return checked.Expr != nil && checked.Expr == consumed.Expr && checked.Invalidated == consumed.Invalidated
}
