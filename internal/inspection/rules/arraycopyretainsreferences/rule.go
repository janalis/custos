// Package arraycopyretainsreferences implements the native ArrayCopyRetainsReferences inspection.
package arraycopyretainsreferences

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Detach referenced array elements before modifying the copy."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayCopyRetainsReferences" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	write := n.(*syntax.Assign)
	if write.ByRef || write.Op.Kind != syntax.TEqual {
		return
	}
	dim, ok := write.Var.(*syntax.ArrayDimFetch)
	if !ok || variable(dim.Var) == "" {
		return
	}
	stmt, ok := write.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, stmt)
	if !ok {
		return
	}
	copyStmt, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	copy, ok := copyStmt.Expr.(*syntax.Assign)
	if !ok || copy.ByRef || copy.Op.Kind != syntax.TEqual || variable(copy.Var) != variable(dim.Var) || variable(copy.Value) == "" || variable(copy.Var) == variable(copy.Value) {
		return
	}
	prev, ok = astquery.PrevStmtNoDoc(ctx.File, copyStmt)
	if !ok {
		return
	}
	refStmt, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	ref, ok := refStmt.Expr.(*syntax.Assign)
	if !ok || !ref.ByRef || variable(ref.Var) == "" {
		return
	}
	element, ok := ref.Value.(*syntax.ArrayDimFetch)
	if !ok || variable(element.Var) != variable(copy.Value) {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, dim.Dim)
	bound, bk := semanticquery.NativeArrayKey(ctx, element.Dim)
	if !known || !bk || key != bound {
		return
	}
	prev, ok = astquery.PrevStmtNoDoc(ctx.File, refStmt)
	if !ok {
		return
	}
	initStmt, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	initial, ok := initStmt.Expr.(*syntax.Assign)
	if !ok || initial.ByRef || variable(initial.Var) != variable(copy.Value) || semanticquery.NativeArray(ctx, initial.Value) == nil {
		return
	}
	read, ok := echoExpr(ctx, stmt).(*syntax.ArrayDimFetch)
	if !ok || variable(read.Var) != variable(copy.Value) {
		return
	}
	observed, rk := semanticquery.NativeArrayKey(ctx, read.Dim)
	if rk && observed == key {
		ctx.ReportNode(write, message)
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
