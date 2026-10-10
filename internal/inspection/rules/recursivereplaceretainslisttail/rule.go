// Package recursivereplaceretainslisttail implements the native RecursiveReplaceRetainsListTail inspection.
package recursivereplaceretainslisttail

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Replace the list explicitly to remove its old tail."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RecursiveReplaceRetainsListTail" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_replace_recursive") {
		return
	}
	a, ak := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	b, bk := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 1, "replacements"))
	if !ak || !bk {
		return
	}
	assignment, ok := call.Parent().(*syntax.Assign)
	if !ok || assignment.ByRef {
		return
	}
	stmt, ok := assignment.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	outer, ok := echoExpr(ctx, stmt).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	inner, ok := outer.Var.(*syntax.ArrayDimFetch)
	if !ok || variable(inner.Var) == "" || variable(inner.Var) != variable(assignment.Var) {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, inner.Dim)
	index, ik := semanticquery.NativeInt(ctx, outer.Dim)
	if !known || !ik {
		return
	}
	left, lok := a[key]
	right, rok := b[key]
	if !lok || !rok {
		return
	}
	l := semanticquery.NativeArray(ctx, left)
	r := semanticquery.NativeArray(ctx, right)
	if !list(l) || !list(r) {
		return
	}
	if index >= int64(len(r.Items)) && index < int64(len(l.Items)) {
		ctx.ReportNode(call, message)
	}
}

func echoExpr(ctx *analysis.Context, s syntax.Stmt) syntax.Expr {
	next, ok := astquery.NextStmt(ctx.File, s)
	if !ok {
		return nil
	}
	echo, ok := next.(*syntax.Echo)
	if !ok || len(echo.Exprs) != 1 {
		return nil
	}
	return syntax.UnwrapParens(echo.Exprs[0])
}

func variable(e syntax.Expr) string {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return ""
	}
	return v.Name
}

func list(a *syntax.Array) bool {
	if a == nil {
		return false
	}
	for _, item := range a.Items {
		if item.Key != nil {
			return false
		}
	}
	return true
}
