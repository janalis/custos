// Package scandirfailureiterated implements the native ScandirFailureIterated inspection.
package scandirfailureiterated

import (
	"custos/internal/inspection/analysis"
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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	loop := n.(*syntax.Foreach)
	if !semanticquery.ExpansionDUnaliased(ctx, loop.Expr, n) && !semanticquery.NativeFailureFlowOnly(ctx, loop.Expr) {
		return
	}
	c, ok := semanticquery.NativeLocalBuiltinResult(ctx, loop.Expr, "false", "scandir").(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, c, "scandir") && (semanticquery.NativeFailureFlowOnly(ctx, loop.Expr) || !semanticquery.NativeLocalSentinelGuard(ctx, loop.Expr, "false")) && !ctx.Flow().Excludes(loop.Expr, "false") && (semanticquery.NativeFailureFlowOnly(ctx, loop.Expr) || !narrowedNotFalse(ctx, loop.Expr)) {
		ctx.ReportNode(loop.Expr, message)
	}
}

// narrowedNotFalse reports whether the conditions guarding the iterated
// local leave its native type known and without false. A static local is
// left to the flow analysis, which knows the calls that may enter the
// function again and assign it.
func narrowedNotFalse(ctx *analysis.Context, value syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(value).(*syntax.Variable)
	if !ok || staticLocal(v) {
		return false
	}
	t := ctx.Types().Native().TypeOf(value)
	return !t.IsUnknown() && !t.Has("false")
}

// staticLocal reports whether v is declared `static` in its scope.
func staticLocal(v *syntax.Variable) bool {
	scope := syntax.EnclosingVariableScope(v)
	if scope == nil {
		return false // no static locals at file level
	}
	found := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return n == scope
		case *syntax.StaticStmt:
			for _, sv := range x.Vars {
				found = found || sv.Var != nil && sv.Var.Name == v.Name
			}
		}
		return !found
	})
	return found
}
