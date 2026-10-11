package intlformattingfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard ties success checks to one format result: a test only counts when it
// reads the same variable while that variable still holds the same format call.
type guard struct {
	ctx    *analysis.Context
	input  syntax.Expr
	origin *syntax.MethodCall
}

// protects reports whether a check on the format result dominates the use:
// an enclosing branch, ternary arm or short-circuit operand, or an earlier
// sibling statement that leaves the path when formatting failed.
func (g guard) protects() bool {
	at := g.input.Span()
	for p := g.input.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			if g.branch(x) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(at) && g.proves(x.Cond, true) {
				return true
			}
			if x.Else.Span().Contains(at) && g.proves(x.Cond, false) {
				return true
			}
		case *syntax.Binary:
			if and, or := connective(x); (and || or) && x.Right.Span().Contains(at) && g.proves(x.Left, and) {
				return true
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, idx, _ := astquery.StmtList(g.ctx.File, st)
		for i := 0; i < idx; i++ {
			if g.leaves(list[i]) {
				return true
			}
		}
	}
	return false
}

// branch handles a use inside one arm of an if chain: the arm's own condition
// held, or an earlier condition of the chain failed.
func (g guard) branch(x *syntax.If) bool {
	at := g.input.Span()
	if x.Body.Span().Contains(at) {
		return g.proves(x.Cond, true)
	}
	if x.Cond.Span().Contains(at) {
		return false
	}
	if g.proves(x.Cond, false) {
		return true
	}
	for _, ei := range x.ElseIfs {
		if ei.Body.Span().Contains(at) {
			return g.proves(ei.Cond, true)
		}
		if ei.Cond.Span().Contains(at) {
			return false
		}
		if g.proves(ei.Cond, false) {
			return true
		}
	}
	return false
}

// leaves reports whether control only continues past st with formatted text.
func (g guard) leaves(st syntax.Stmt) bool {
	switch x := st.(type) {
	case *syntax.If:
		if syntax.Terminates(x.Body) && g.proves(x.Cond, false) {
			return true
		}
		return len(x.ElseIfs) == 0 && x.Else != nil && syntax.Terminates(x.Else.Body) && g.proves(x.Cond, true)
	case *syntax.ExprStmt:
		if e, ok := syntax.UnwrapParens(x.Expr).(*syntax.Binary); ok {
			and, or := connective(e)
			_, throws := syntax.UnwrapParens(e.Right).(*syntax.Throw)
			return (and || or) && (throws || syntax.ExitInvocation(e.Right)) && g.proves(e.Left, or)
		}
	}
	return false
}

// proves reports whether cond evaluating to truth rules out a failed format.
// The result is a string or false, so a truthy value is always a string.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.proves(x.Expr, !truth)
	case *syntax.Binary:
		if and, or := connective(x); and || or {
			l, r := g.proves(x.Left, truth), g.proves(x.Right, truth)
			if and == truth {
				return l || r
			}
			return l && r
		}
		switch x.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.againstFalse(x)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.againstFalse(x)
		}
		return false
	case *syntax.Empty:
		return !truth && g.same(x.Expr)
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(g.ctx, x) {
		case "is_string":
			return truth && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
		case "is_bool":
			return !truth && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
		}
		return false
	default:
		return truth && g.same(x)
	}
}

func (g guard) againstFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if value, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok && !value && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same accepts the checked variable while it still holds the format result,
// and the assignment that stores that result when it is tested in place.
func (g guard) same(e syntax.Expr) bool {
	if e == nil {
		return false
	}
	e = syntax.UnwrapParens(e)
	if a, ok := e.(*syntax.Assign); ok {
		return astquery.Equivalent(g.ctx.File, a.Var, g.input) && syntax.UnwrapParens(a.Value) == syntax.Expr(g.origin)
	}
	return astquery.Equivalent(g.ctx.File, e, g.input) && semanticquery.NativeValue(g.ctx, e) == syntax.Expr(g.origin)
}

func connective(b *syntax.Binary) (and, or bool) {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return true, false
	case syntax.TBooleanOr, syntax.TOr:
		return false, true
	}
	return false, false
}
