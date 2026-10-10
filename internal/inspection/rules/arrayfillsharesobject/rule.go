// Package arrayfillsharesobject implements the native ArrayFillSharesObject inspection.
package arrayfillsharesobject

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Construct a separate object for each array entry."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFillSharesObject" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_fill") {
		return
	}
	creation, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 2, "value")).(*syntax.New)
	if !ok {
		return
	}
	name, ok := creation.Class.(*syntax.Name)
	if !ok || ctx.Index().Class(ctx.Names().Class(name.Value, name.Span().Start), ctx.PHP) == nil {
		return
	}
	count, ck := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "count"))
	start, sk := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 0, "start_index"))
	if !ck || !sk || count <= 1 || start < 0 || start > 1<<62 || count > 1<<62 {
		return
	}
	a, ok := call.Parent().(*syntax.Assign)
	if !ok || a.ByRef || variable(a.Var) == "" {
		return
	}
	stmt, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, stmt)
	if !ok {
		return
	}
	writeStmt, ok := next.(*syntax.ExprStmt)
	if !ok {
		return
	}
	write, ok := writeStmt.Expr.(*syntax.Assign)
	if !ok || write.ByRef {
		return
	}
	prop, ok := write.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	dim, ok := prop.Var.(*syntax.ArrayDimFetch)
	if !ok || variable(dim.Var) != variable(a.Var) {
		return
	}
	index, ik := semanticquery.NativeInt(ctx, dim.Dim)
	if !ik || index < start || index >= start+count {
		return
	}
	read, ok := echoExpr(ctx, writeStmt).(*syntax.PropertyFetch)
	if !ok {
		return
	}
	other, ok := read.Var.(*syntax.ArrayDimFetch)
	if !ok || variable(other.Var) != variable(a.Var) {
		return
	}
	second, jk := semanticquery.NativeInt(ctx, other.Dim)
	p, pk := prop.Name.(*syntax.Identifier)
	q, qk := read.Name.(*syntax.Identifier)
	if jk && pk && qk && p.Value == q.Value && second >= start && second < start+count && second != index {
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
