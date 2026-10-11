package xsltransformationfailureconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard carries the consumed expression and the transformation call it holds.
type guard struct {
	ctx   *analysis.Context
	input syntax.Expr
	call  *syntax.MethodCall
}

// dominates reports whether a truthiness, loose or type test of the same
// value cannot let false reach the input.
func (g guard) dominates() bool {
	at := g.input.Span()
	for p := g.input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			if x.Body.Span().Contains(at) && g.rejects(x.Cond, true) {
				return true
			}
			if x.Else != nil && x.Else.Body.Span().Contains(at) && g.elseRejects(x) {
				return true
			}
		case *syntax.ElseIf:
			if x.Body.Span().Contains(at) && g.rejects(x.Cond, true) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(at) && g.rejects(x.Cond, true) {
				return true
			}
			if x.Else.Span().Contains(at) && g.rejects(x.Cond, false) {
				return true
			}
		case *syntax.Binary:
			// The right operand runs only when the left one passed (&&) or failed (||).
			if (conjunction(x) || disjunction(x)) && x.Right.Span().Contains(at) && g.rejects(x.Left, conjunction(x)) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && g.exitedBefore(st) {
			return true
		}
	}
	return false
}

// elseRejects holds when any condition of the chain is false only for a
// value other than false: the else branch runs after all of them failed.
func (g guard) elseRejects(branch *syntax.If) bool {
	if g.rejects(branch.Cond, false) {
		return true
	}
	for _, alt := range branch.ElseIfs {
		if g.rejects(alt.Cond, false) {
			return true
		}
	}
	return false
}

// exitedBefore looks for an earlier sibling guard that leaves the scope on
// failure. The same-value proof covers anything written in between.
func (g guard) exitedBefore(st syntax.Stmt) bool {
	for prev, ok := astquery.PrevStmt(g.ctx.File, st); ok; prev, ok = astquery.PrevStmt(g.ctx.File, prev) {
		branch, isIf := prev.(*syntax.If)
		if isIf && branch.Else == nil && len(branch.ElseIfs) == 0 && syntax.Terminates(branch.Body) && g.rejects(branch.Cond, false) {
			return true
		}
	}
	return false
}

// rejects reports whether cond evaluating to truth proves the value is not
// false.
func (g guard) rejects(cond syntax.Expr, truth bool) bool {
	cond = syntax.UnwrapParens(cond)
	switch c := cond.(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && g.rejects(c.Expr, !truth)
	case *syntax.Empty:
		return !truth && g.same(c.Expr)
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, c, "is_string") && g.same(semanticquery.CallArgument(c.Args, 0, "value"))
	case *syntax.Binary:
		switch {
		case conjunction(c):
			return truth && (g.rejects(c.Left, true) || g.rejects(c.Right, true))
		case disjunction(c):
			return !truth && (g.rejects(c.Left, false) || g.rejects(c.Right, false))
		case c.Op.Kind == syntax.TIsEqual, c.Op.Kind == syntax.TIsIdentical:
			return !truth && g.comparesFalse(c)
		case c.Op.Kind == syntax.TIsNotEqual, c.Op.Kind == syntax.TIsNotIdentical:
			return truth && g.comparesFalse(c)
		}
		return false
	}
	return truth && g.same(cond)
}

func conjunction(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func disjunction(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}

func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if value, known := astquery.BoolConst(syntax.UnwrapParens(pair[1])); known && !value && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same accepts the unchanged value, or the call assigned inside the test.
func (g guard) same(e syntax.Expr) bool {
	e = syntax.UnwrapParens(e)
	if a, ok := e.(*syntax.Assign); ok {
		return a.Op.Kind == syntax.TEqual && !a.ByRef && syntax.UnwrapParens(a.Value) == syntax.Expr(g.call)
	}
	return semanticquery.NativeSameValue(g.ctx, g.input, e)
}
