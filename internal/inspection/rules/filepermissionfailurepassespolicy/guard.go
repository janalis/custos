package filepermissionfailurepassespolicy

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// failureRejected reports whether an earlier test of the same lookup result
// already keeps false away from e: a short-circuit operand, an enclosing
// branch, or a preceding sibling guard whose body leaves the block.
func failureRejected(ctx *analysis.Context, e syntax.Expr) bool {
	var child syntax.Node = e
	for p := e.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Binary:
			if x.Right == child && (x.Op.Kind == syntax.TBooleanAnd && excludesFalse(ctx, e, x.Left, true) || x.Op.Kind == syntax.TBooleanOr && excludesFalse(ctx, e, x.Left, false)) {
				return true
			}
		case *syntax.If:
			if x.Body == child && excludesFalse(ctx, e, x.Cond, true) {
				return true
			}
		case *syntax.ElseIf:
			if x.Body == child && excludesFalse(ctx, e, x.Cond, true) || earlierBranchesExclude(ctx, e, x) {
				return true
			}
		case *syntax.Else:
			if earlierBranchesExclude(ctx, e, x) {
				return true
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, _ := astquery.StmtList(ctx.File, st)
		for _, previous := range list[:max(i, 0)] {
			guard, ok := previous.(*syntax.If)
			if ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && excludesFalse(ctx, e, guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

// earlierBranchesExclude handles an elseif or else arm: every condition
// written before it evaluated to false when the arm runs.
func earlierBranchesExclude(ctx *analysis.Context, e syntax.Expr, arm syntax.Node) bool {
	owner := arm.Parent().(*syntax.If)
	if excludesFalse(ctx, e, owner.Cond, false) {
		return true
	}
	for _, other := range owner.ElseIfs {
		if other == arm {
			break
		}
		if excludesFalse(ctx, e, other.Cond, false) {
			return true
		}
	}
	return false
}

// excludesFalse reports whether cond evaluating to truth proves that the
// value read by e is not false. Only tests of that same reaching value count.
func excludesFalse(ctx *analysis.Context, e, cond syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Variable:
		return truth && semanticquery.NativeSameValue(ctx, e, c)
	case *syntax.Empty:
		return !truth && semanticquery.NativeSameValue(ctx, e, c.Expr)
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && excludesFalse(ctx, e, c.Expr, !truth)
	case *syntax.FuncCall:
		if c.Args == nil || len(c.Args.Args) != 1 {
			return false
		}
		arg, ok := c.Args.Args[0].(*syntax.Arg)
		if !ok || arg.Name != nil || arg.Unpack || !semanticquery.NativeSameValue(ctx, e, arg.Value) {
			return false
		}
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "is_int", "is_integer", "is_long":
			return truth
		case "is_bool":
			return !truth
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd:
			return truth && (excludesFalse(ctx, e, c.Left, true) || excludesFalse(ctx, e, c.Right, true))
		case syntax.TBooleanOr:
			return !truth && (excludesFalse(ctx, e, c.Left, false) || excludesFalse(ctx, e, c.Right, false))
		case syntax.TIsEqual, syntax.TIsIdentical, syntax.TIsNotEqual, syntax.TIsNotIdentical:
			// Loose and strict forms agree here: a value that is not loosely
			// equal to false is truthy, hence not false either.
			if truth != (c.Op.Kind == syntax.TIsNotEqual || c.Op.Kind == syntax.TIsNotIdentical) {
				return false
			}
			for _, pair := range [][2]syntax.Expr{{c.Left, c.Right}, {c.Right, c.Left}} {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				if ok && strings.EqualFold(semanticquery.GlobalConstName(ctx, constant), "false") && semanticquery.NativeSameValue(ctx, e, pair[0]) {
					return true
				}
			}
		}
	}
	return false
}
