// Package failedclosurebindinginvoked implements the native FailedClosureBindingInvoked inspection.
package failedclosurebindinginvoked

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check closure binding before invoking it."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FailedClosureBindingInvoked" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	bound := semanticquery.NativeLocalValue(ctx, call.Name)
	var closure, object syntax.Expr
	switch b := bound.(type) {
	case *syntax.MethodCall:
		if !semanticquery.NativeMethod(ctx, b, "Closure", "bindTo") {
			return
		}
		closure = b.Var
		object = semanticquery.CallArgument(b.Args, 0, "newThis")
	case *syntax.StaticCall:
		cl, ok := b.Class.(*syntax.Name)
		name, ok2 := b.Name.(*syntax.Identifier)
		if !ok || !ok2 || !strings.EqualFold(ctx.Names().Class(cl.Value, cl.Span().Start), "Closure") || !strings.EqualFold(name.Value, "bind") {
			return
		}
		closure = semanticquery.CallArgument(b.Args, 0, "closure")
		object = semanticquery.CallArgument(b.Args, 1, "newThis")
	default:
		return
	}
	static := false
	switch c := semanticquery.NativeValue(ctx, closure).(type) {
	case *syntax.Closure:
		static = c.Static
	case *syntax.ArrowFunction:
		static = c.Static
	}
	if !static || object == nil || !provenObject(ctx, object) || guarded(ctx, call.Name) {
		return
	}
	ctx.ReportNode(call, message)
}

// provenObject requires every alternative of the bound value to be a class
// instance: binding null to a static closure succeeds.
func provenObject(ctx *analysis.Context, object syntax.Expr) bool {
	t := ctx.Types().Native().TypeOf(object)
	return len(t.Classes()) > 0 && len(t.Classes()) == len(t.Atoms())
}

// guarded reports whether the invoked binding result is known to be non-null
// where it is called: inside a branch, conditional arm or short-circuit
// operand selected by a null check, or after a terminating one.
func guarded(ctx *analysis.Context, callee syntax.Expr) bool {
	in := func(n syntax.Node) bool { return n != nil && n.Span().Contains(callee.Span()) }
	for p := callee.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Ternary:
			if in(x.Then) && nonNull(ctx, callee, x.Cond, true) || in(x.Else) && nonNull(ctx, callee, x.Cond, false) {
				return true
			}
		case *syntax.Binary:
			and := x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd
			or := x.Op.Kind == syntax.TBooleanOr || x.Op.Kind == syntax.TOr
			if (and || or) && in(x.Right) && nonNull(ctx, callee, x.Left, and) {
				return true
			}
		case *syntax.If:
			if branchGuarded(ctx, callee, x, in) {
				return true
			}
		}
		st, ok := p.(syntax.Stmt)
		for ok {
			if st, ok = astquery.PrevStmt(ctx.File, st); !ok {
				break
			}
			if guard, isIf := st.(*syntax.If); isIf && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && nonNull(ctx, callee, guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

// branchGuarded reports whether the if branch holding the invocation is only
// entered with a non-null result: its own condition passed, or one tested
// before it failed.
func branchGuarded(ctx *analysis.Context, callee syntax.Expr, x *syntax.If, in func(syntax.Node) bool) bool {
	conds, bodies := []syntax.Expr{x.Cond}, []syntax.Stmt{x.Body}
	for _, branch := range x.ElseIfs {
		conds, bodies = append(conds, branch.Cond), append(bodies, branch.Body)
	}
	earlier := false
	for i, cond := range conds {
		if in(cond) {
			return earlier
		}
		if in(bodies[i]) {
			return earlier || nonNull(ctx, callee, cond, true)
		}
		earlier = earlier || nonNull(ctx, callee, cond, false)
	}
	return earlier
}

// nonNull reports whether cond evaluating to truth proves the binding result
// is not null.
func nonNull(ctx *analysis.Context, callee, cond syntax.Expr, truth bool) bool {
	same := func(e syntax.Expr) bool { return semanticquery.NativeSameValue(ctx, callee, e) }
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && nonNull(ctx, callee, c.Expr, !truth)
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (nonNull(ctx, callee, c.Left, true) || nonNull(ctx, callee, c.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (nonNull(ctx, callee, c.Left, false) || nonNull(ctx, callee, c.Right, false))
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			if truth != (c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual) {
				return false
			}
			for _, pair := range [][2]syntax.Expr{{c.Left, c.Right}, {c.Right, c.Left}} {
				constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
				if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "null") && same(pair[0]) {
					return true
				}
			}
		}
		return false
	case *syntax.Instanceof:
		return truth && same(c.Expr)
	case *syntax.Isset:
		return truth && slices.ContainsFunc(c.Vars, same)
	case *syntax.Empty:
		return !truth && same(c.Expr)
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "is_null":
			return !truth && same(semanticquery.CallArgument(c.Args, 0, "value"))
		case "is_callable", "is_object":
			return truth && same(semanticquery.CallArgument(c.Args, 0, "value"))
		}
		return false
	}
	return truth && same(cond)
}
