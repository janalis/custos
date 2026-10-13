package transliterationfailureconsumed

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard proves that a local holding a transliteration result cannot be false
// where it is consumed. The flow facts are asked first and read truthiness and
// loose tests too; this walk only reads strict false comparisons and
// is_string(), for the shapes the flow leaves open: match (true) arms and a
// condition whose two operands both exclude false.
type guard struct {
	ctx   *analysis.Context
	input *syntax.Variable
}

// dominates walks the enclosing branches of the consumer and the terminating
// guard statement directly before each enclosing statement.
func (g guard) dominates() bool {
	for p := g.input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			if g.within(x.Body) {
				if g.proves(x.Cond, true) {
					return true
				}
			} else if !g.within(x.Cond) && g.failed(x) {
				return true
			}
		case *syntax.ElseIf:
			if g.within(x.Body) && g.proves(x.Cond, true) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && g.within(x.Then) && g.proves(x.Cond, true) {
				return true
			}
			if g.within(x.Else) && g.proves(x.Cond, false) {
				return true
			}
		case *syntax.Binary:
			if g.within(x.Right) && g.shortCircuit(x) {
				return true
			}
		case *syntax.MatchArm:
			if g.arm(x) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok {
			previous, exists := astquery.PrevStmt(g.ctx.File, st)
			if check, ok := previous.(*syntax.If); exists && ok && check.Else == nil && len(check.ElseIfs) == 0 && syntax.Terminates(check.Body) && g.proves(check.Cond, false) {
				return true
			}
		}
	}
	return false
}

func (g guard) within(n syntax.Node) bool { return n.Span().Contains(g.input.Span()) }

// failed handles a consumer in an elseif or else part: the leading condition
// and every elseif condition written before the consumer evaluated to false.
func (g guard) failed(branch *syntax.If) bool {
	if g.proves(branch.Cond, false) {
		return true
	}
	for _, alternative := range branch.ElseIfs {
		if alternative.Span().End <= g.input.Span().Start && g.proves(alternative.Cond, false) {
			return true
		}
	}
	return false
}

// shortCircuit handles a consumer in the right operand, which is evaluated
// only when the left operand was true (&&, and) or false (||, or).
func (g guard) shortCircuit(b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return g.proves(b.Left, true)
	case syntax.TBooleanOr, syntax.TOr:
		return g.proves(b.Left, false)
	}
	return false
}

// arm handles match (true): an arm is reached only after every condition of
// the earlier arms failed, and its body only when its own condition held.
func (g guard) arm(arm *syntax.MatchArm) bool {
	m, ok := arm.Parent().(*syntax.Match)
	if !ok || !constant(m.Cond, "true") {
		return false
	}
	if len(arm.Conds) == 1 && g.within(arm.Body) && g.proves(arm.Conds[0], true) {
		return true
	}
	for _, earlier := range m.Arms {
		if earlier == arm {
			break
		}
		for _, cond := range earlier.Conds {
			if g.proves(cond, false) {
				return true
			}
		}
	}
	return false
}

// proves reports whether cond evaluating to truth excludes false for the input.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.proves(x.Expr, !truth)
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, x, "is_string") && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			// A true conjunction held on both sides; a false one failed on either.
			if truth {
				return g.proves(x.Left, true) || g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) && g.proves(x.Right, false)
		case syntax.TBooleanOr, syntax.TOr:
			if truth {
				return g.proves(x.Left, true) && g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) || g.proves(x.Right, false)
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			if truth != (x.Op.Kind == syntax.TIsNotIdentical) {
				return false
			}
			return (constant(x.Right, "false") && g.same(x.Left)) || (constant(x.Left, "false") && g.same(x.Right))
		}
	}
	return false
}

// same ties a checked operand to the value reaching the consumer. A plain
// assignment to the variable is the value it stores; a rebinding between the
// check and the consumer changes the reaching producer and removes the proof.
func (g guard) same(e syntax.Expr) bool {
	e = syntax.UnwrapParens(e)
	want := g.ctx.Flow().Value(g.input)
	if assign, ok := e.(*syntax.Assign); ok {
		return assign.Op.Kind == syntax.TEqual && !assign.ByRef && astquery.Equivalent(g.ctx.File, assign.Var, g.input) &&
			syntax.UnwrapParens(assign.Value) == syntax.UnwrapParens(want.Expr)
	}
	if !astquery.Equivalent(g.ctx.File, e, g.input) {
		return false
	}
	got := g.ctx.Flow().Value(e)
	return got.Expr == want.Expr && got.Invalidated == want.Invalidated
}

func constant(e syntax.Expr, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), name)
}
