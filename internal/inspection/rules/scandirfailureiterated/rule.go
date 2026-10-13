// Package scandirfailureiterated implements the native ScandirFailureIterated inspection.
package scandirfailureiterated

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check directory enumeration before iterating."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ScandirFailureIterated" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	loop := n.(*syntax.Foreach)
	if !semanticquery.ExpansionDUnaliased(ctx, loop.Expr, n) {
		return
	}
	c, ok := semanticquery.NativeLocalResult(ctx, loop.Expr, "false").(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, c, "scandir") && !semanticquery.NativeLocalSentinelGuard(ctx, loop.Expr, "false") && !ctx.Flow().Excludes(loop.Expr, "false") && (semanticquery.NativeLabelled(ctx, loop.Expr) || !guarded(ctx, loop.Expr)) {
		ctx.ReportNode(loop.Expr, message)
	}
}

// guarded proves that a branch condition or an earlier terminating guard on
// the current path rejects false for the iterated local.
func guarded(ctx *analysis.Context, value syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(value).(*syntax.Variable)
	if !ok {
		return false
	}
	start := value.Span().Start
	for p := value.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.If:
			// Past the main body means an elseif or else branch.
			if rejectsFalse(ctx, value, x.Cond, x.Body.Span().Contains(value.Span())) {
				return true
			}
			for _, branch := range x.ElseIfs {
				if branch.Body.Span().End <= start && rejectsFalse(ctx, value, branch.Cond, false) {
					return true
				}
			}
		case *syntax.ElseIf:
			if rejectsFalse(ctx, value, x.Cond, true) {
				return true
			}
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			// A guard outside a loop proves nothing once the loop rewrites the local.
			if assigns(p, variable.Name) {
				return false
			}
		}
		// Every ancestor below the scope is a statement.
		list, i, _ := astquery.StmtList(ctx.File, p.(syntax.Stmt))
		for _, previous := range list[:max(i, 0)] {
			guard, ok := previous.(*syntax.If)
			if !ok {
				continue
			}
			if syntax.Terminates(guard.Body) && rejectsFalse(ctx, value, guard.Cond, false) {
				return true
			}
			if guard.Else != nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Else.Body) && rejectsFalse(ctx, value, guard.Cond, true) {
				return true
			}
		}
	}
	return false
}

func assigns(loop syntax.Node, name string) bool {
	found := false
	syntax.Inspect(loop, func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			v, ok := a.Var.(*syntax.Variable)
			found = found || ok && v.Name == name
		}
		return !found
	})
	return found
}

// rejectsFalse reports whether cond evaluating to truth proves value is not
// false. A conjunction proves it through either operand only when it holds,
// a disjunction only when it fails; the opposite outcome needs both operands.
func rejectsFalse(ctx *analysis.Context, value, cond syntax.Expr, truth bool) bool {
	cond = syntax.UnwrapParens(cond)
	switch x := cond.(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && rejectsFalse(ctx, value, x.Expr, !truth)
	case *syntax.Empty:
		return !truth && semanticquery.NativeSameValue(ctx, value, x.Expr)
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(ctx, x, "is_array") && semanticquery.NativeSameValue(ctx, value, semanticquery.CallArgument(x.Args, 0, "value"))
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			left, right := rejectsFalse(ctx, value, x.Left, truth), rejectsFalse(ctx, value, x.Right, truth)
			if truth == (x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd) {
				return left || right
			}
			return left && right
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && comparesFalse(ctx, value, x)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && comparesFalse(ctx, value, x)
		}
		return false
	}
	return truth && semanticquery.NativeSameValue(ctx, value, cond)
}

func comparesFalse(ctx *analysis.Context, value syntax.Expr, b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		c, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") && semanticquery.NativeSameValue(ctx, value, pair[0]) {
			return true
		}
	}
	return false
}
