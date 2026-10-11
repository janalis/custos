package pregreplacefailureunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// nullGuarded proves that a null test on the same reaching value dominates
// input: an enclosing branch, ternary arm or short-circuit operand, or an
// earlier terminating guard in an enclosing statement list.
func nullGuarded(ctx *analysis.Context, input syntax.Expr) bool {
	for child, p := syntax.Node(input), input.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch p := p.(type) {
		case *syntax.If:
			// Every elseif and else arm runs only after the first condition failed.
			if child != syntax.Node(p.Cond) && excludesNull(ctx, input, p.Cond, child == syntax.Node(p.Body)) {
				return true
			}
		case *syntax.ElseIf:
			if child == syntax.Node(p.Body) && excludesNull(ctx, input, p.Cond, true) {
				return true
			}
		case *syntax.Ternary:
			if child != syntax.Node(p.Cond) && excludesNull(ctx, input, p.Cond, child == syntax.Node(p.Then)) {
				return true
			}
		case *syntax.Binary:
			and, or := logicalOperator(p)
			if (and || or) && child == syntax.Node(p.Right) && excludesNull(ctx, input, p.Left, and) {
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
			if ok && syntax.Terminates(guard.Body) && excludesNull(ctx, input, guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

func logicalOperator(b *syntax.Binary) (and, or bool) {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return true, false
	case syntax.TBooleanOr, syntax.TOr:
		return false, true
	}
	return false, false
}

// excludesNull reports whether condition having the given truth proves that
// input is not null. General truthiness and loose comparisons prove nothing.
func excludesNull(ctx *analysis.Context, input, condition syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(condition).(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && excludesNull(ctx, input, c.Expr, !truth)
	case *syntax.Binary:
		and, or := logicalOperator(c)
		switch {
		case and:
			return truth && (excludesNull(ctx, input, c.Left, true) || excludesNull(ctx, input, c.Right, true))
		case or:
			return !truth && (excludesNull(ctx, input, c.Left, false) || excludesNull(ctx, input, c.Right, false))
		case c.Op.Kind == syntax.TIsNotIdentical:
			return truth && comparesNull(ctx, input, c)
		case c.Op.Kind == syntax.TIsIdentical:
			return !truth && comparesNull(ctx, input, c)
		}
	case *syntax.Isset:
		for _, v := range c.Vars {
			if truth && sameValue(ctx, input, v) {
				return true
			}
		}
	case *syntax.FuncCall:
		tested := semanticquery.CallArgument(c.Args, 0, "value")
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "is_null":
			return !truth && sameValue(ctx, input, tested)
		case "is_string":
			return truth && sameValue(ctx, input, tested)
		}
	}
	return false
}

func comparesNull(ctx *analysis.Context, input syntax.Expr, b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "null") && sameValue(ctx, input, pair[0]) {
			return true
		}
	}
	return false
}

// sameValue ties a guard to the reaching producer: a reassignment or a
// repeated call is a new value even when the tested expression reads the same.
func sameValue(ctx *analysis.Context, input, tested syntax.Expr) bool {
	if tested == nil {
		return false
	}
	input, tested = syntax.UnwrapParens(input), syntax.UnwrapParens(tested)
	// `null !== $v = call()` tests the value the assignment stores.
	if assign, ok := tested.(*syntax.Assign); ok && assign.Op.Kind == syntax.TEqual && !assign.ByRef {
		stored := ctx.Flow().Value(input).Expr
		return stored != nil && astquery.Equivalent(ctx.File, input, assign.Var) && syntax.UnwrapParens(stored) == syntax.UnwrapParens(assign.Value)
	}
	if !astquery.Equivalent(ctx.File, input, tested) {
		return false
	}
	a, b := ctx.Flow().Value(input), ctx.Flow().Value(tested)
	return a.Expr == b.Expr && a.Invalidated == b.Invalidated
}
