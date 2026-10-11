package filereadfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard proves failure handling that type narrowing does not model: loose
// comparisons with a falsy constant, alternatives that never return and
// branches ending in a call declared never.
type guard struct {
	ctx  *analysis.Context
	read *syntax.FuncCall
}

// guarded reports whether every path from the read to use excludes false.
func guarded(ctx *analysis.Context, use syntax.Expr, read *syntax.FuncCall) bool {
	g := guard{ctx, read}
	if g.alternativeNeverReturns(use) {
		return true
	}
	var child syntax.Node = use
	for p := use.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			if child == x.Right && (logicalAnd(x) || logicalOr(x)) && g.excludes(x.Left, logicalAnd(x)) {
				return true
			}
		case *syntax.Ternary:
			if child != x.Cond && g.excludes(x.Cond, child == x.Then) {
				return true
			}
		case *syntax.If:
			if child == x.Body && g.excludes(x.Cond, true) {
				return true
			}
		case *syntax.ElseIf, *syntax.Else:
			// Reached only after the owning condition evaluated to false.
			if owner, ok := p.Parent().(*syntax.If); ok && g.excludes(owner.Cond, false) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && g.earlierExit(st) {
			return true
		}
	}
	return false
}

func logicalAnd(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func logicalOr(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}

// alternativeNeverReturns recognizes `$text = read() or die()`: the tracked
// assignment is the left operand of a disjunction whose other side never
// returns, so any later use holds a truthy result.
func (g guard) alternativeNeverReturns(use syntax.Expr) bool {
	assign, ok := parentPastParens(g.read).(*syntax.Assign)
	if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual || syntax.UnwrapParens(assign.Value) != syntax.Expr(g.read) {
		return false
	}
	or, ok := parentPastParens(assign).(*syntax.Binary)
	return ok && logicalOr(or) && syntax.UnwrapParens(or.Left) == syntax.Expr(assign) &&
		!or.Span().Contains(use.Span()) && g.neverReturns(or.Right)
}

func parentPastParens(n syntax.Node) syntax.Node {
	p := n.Parent()
	for {
		paren, ok := p.(*syntax.Paren)
		if !ok {
			return p
		}
		p = paren.Parent()
	}
}

// earlierExit finds an earlier sibling `if` without alternatives whose body
// leaves the path whenever its condition does not exclude false.
func (g guard) earlierExit(st syntax.Stmt) bool {
	list, i, _ := astquery.StmtList(g.ctx.File, st)
	for _, previous := range list[:max(i, 0)] {
		branch, ok := previous.(*syntax.If)
		if ok && branch.Else == nil && len(branch.ElseIfs) == 0 && g.exits(branch.Body) && g.excludes(branch.Cond, false) {
			return true
		}
	}
	return false
}

func (g guard) exits(body syntax.Stmt) bool {
	if syntax.Terminates(body) {
		return true
	}
	stmts, ok := astquery.BlockStmts(body)
	if !ok {
		stmts = []syntax.Stmt{body}
	}
	for _, st := range stmts {
		if e, ok := st.(*syntax.ExprStmt); ok && g.neverReturns(e.Expr) {
			return true
		}
	}
	return false
}

// neverReturns covers exit, throw expressions and calls declared never.
func (g guard) neverReturns(e syntax.Expr) bool {
	return syntax.ExitInvocation(e) || g.ctx.TypeOf(syntax.UnwrapParens(e)).OnlyOf("never")
}

// excludes reports whether cond evaluating to truth proves the read result is
// not false.
func (g guard) excludes(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.excludes(x.Expr, !truth)
	case *syntax.Empty:
		return !truth && g.same(x.Expr)
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, x, "is_string") && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (g.excludes(x.Left, true) || g.excludes(x.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (g.excludes(x.Left, false) || g.excludes(x.Right, false))
		case syntax.TIsEqual, syntax.TIsIdentical:
			return !truth && g.comparesFalse(x)
		case syntax.TIsNotEqual, syntax.TIsNotIdentical:
			return truth && g.comparesFalse(x)
		}
		return false
	}
	return truth && g.same(cond)
}

// comparesFalse matches the read result against literal false, or, for a loose
// comparison, against any constant that false equals.
func (g guard) comparesFalse(b *syntax.Binary) bool {
	strict := b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsNotIdentical
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if !g.same(pair[0]) {
			continue
		}
		if value, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok {
			return !value
		}
		truth, known := semanticquery.NativeTruth(g.ctx, pair[1])
		return !strict && known && !truth
	}
	return false
}

// same reports whether e holds the tracked read result: the assignment that
// stores it, or a variable whose only reaching value is that earlier read.
func (g guard) same(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Assign:
		return !x.ByRef && x.Op.Kind == syntax.TEqual && syntax.UnwrapParens(x.Value) == syntax.Expr(g.read)
	case *syntax.Variable:
		return g.read.Span().End <= x.Span().Start && semanticquery.NativeValue(g.ctx, x) == syntax.Expr(g.read)
	}
	return false
}
