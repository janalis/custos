package realpathfailureusedaspath

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// guard ties failure checks to one realpath call: a test counts only when the
// tested variable still holds that call's result.
type guard struct {
	ctx    *analysis.Context
	source syntax.Expr
}

// protects reports whether every path reaching at has rejected a false result.
func (g guard) protects(at syntax.Node) bool {
	for child, p := at, at.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			if child == x.Body && g.implies(x.Cond, true) {
				return true
			}
			// Later branches run only after every earlier condition failed.
			if child != x.Body && child != x.Cond && g.implies(x.Cond, false) {
				return true
			}
			for _, branch := range x.ElseIfs {
				if branch.Span().Start < child.Span().Start && g.implies(branch.Cond, false) {
					return true
				}
			}
		case *syntax.ElseIf:
			if child == x.Body && g.implies(x.Cond, true) {
				return true
			}
		case *syntax.While:
			if child == x.Body && g.implies(x.Cond, true) {
				return true
			}
		case *syntax.Ternary:
			if (child == x.Then && g.implies(x.Cond, true)) || (child == x.Else && g.implies(x.Cond, false)) {
				return true
			}
		case *syntax.Binary:
			if child == x.Right && g.shortCircuits(x) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && g.rejectedBefore(st) {
			return true
		}
	}
	return false
}

// shortCircuits reports whether the right operand runs only for a non-false result.
func (g guard) shortCircuits(b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return g.implies(b.Left, true)
	case syntax.TBooleanOr, syntax.TOr:
		return g.implies(b.Left, false)
	}
	return false
}

// rejectedBefore looks for an earlier sibling statement that leaves the block
// when the result is false: a terminating `if`, or `test || throw` / `or die`.
func (g guard) rejectedBefore(st syntax.Stmt) bool {
	list, index, _ := astquery.StmtList(g.ctx.File, st)
	for i := 0; i < index; i++ {
		switch x := list[i].(type) {
		case *syntax.If:
			// Whatever follows the first branch, passing the statement
			// means its condition was false.
			if syntax.Terminates(x.Body) && g.implies(x.Cond, false) {
				return true
			}
		case *syntax.ExprStmt:
			b, ok := syntax.UnwrapParens(x.Expr).(*syntax.Binary)
			if !ok || !g.aborts(b.Right) {
				continue
			}
			// Execution continues only when the left operand skipped the abort.
			switch b.Op.Kind {
			case syntax.TBooleanOr, syntax.TOr:
				if g.implies(b.Left, true) {
					return true
				}
			case syntax.TBooleanAnd, syntax.TAnd:
				if g.implies(b.Left, false) {
					return true
				}
			}
		}
	}
	return false
}

func (g guard) aborts(e syntax.Expr) bool {
	if _, ok := syntax.UnwrapParens(e).(*syntax.Throw); ok {
		return g.ctx.PHP >= phpversion.PHP80 // throw is a statement before 8.0
	}
	return syntax.ExitInvocation(e)
}

// implies reports whether cond evaluating to truth proves the result is not
// false. Conjunctions prove it only when true, disjunctions only when false.
func (g guard) implies(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.implies(x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (g.implies(x.Left, true) || g.implies(x.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (g.implies(x.Left, false) || g.implies(x.Right, false))
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.comparesFalse(x)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.comparesFalse(x)
		}
		return false
	case *syntax.Empty:
		return !truth && g.same(x.Expr)
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, x, "is_string") && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
	}
	return truth && g.same(cond)
}

func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if value, ok := astquery.BoolConst(syntax.UnwrapParens(pair[1])); ok && !value && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same accepts a variable whose reaching value is the tracked call, or the
// assignment that stores it.
func (g guard) same(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Assign:
		return x.Value == g.source
	case *syntax.Variable:
		return semanticquery.NativeLocalValue(g.ctx, x) == g.source
	}
	return false
}
