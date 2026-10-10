// Package gdcolorzerorejectedasfailure implements the native GdColorZeroRejectedAsFailure inspection.
package gdcolorzerorejectedasfailure

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare color allocation failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdColorZeroRejectedAsFailure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.If)
	e := syntax.UnwrapParens(b.Cond)
	var operand syntax.Expr
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		operand = u.Expr
	}
	if p, ok := e.(*syntax.Binary); ok && p.Op.Kind == syntax.TIsEqual {
		for _, pair := range [][2]syntax.Expr{{p.Left, p.Right}, {p.Right, p.Left}} {
			v, k := astquery.BoolConst(syntax.UnwrapParens(pair[1]))
			if k && !v {
				operand = pair[0]
			}
		}
	}
	if operand == nil {
		return
	}
	if a, ok := syntax.UnwrapParens(operand).(*syntax.Assign); ok {
		operand = a.Value
	}
	c, ok := semanticquery.NativeLocalValue(ctx, operand).(*syntax.FuncCall)
	if !ok || (!semanticquery.NativeBuiltin(ctx, c, "imagecolorallocate") && !semanticquery.NativeBuiltin(ctx, c, "imagecolorallocatealpha")) {
		return
	}
	if failure(b.Body) {
		ctx.ReportNode(b.Cond, message)
	}
}

func failure(body syntax.Stmt) bool {
	block, ok := body.(*syntax.Block)
	if !ok || len(block.Stmts) != 1 {
		return false
	}
	switch x := block.Stmts[0].(type) {
	case *syntax.Return:
		v, k := astquery.BoolConst(x.Expr)
		return k && !v
	case *syntax.ExprStmt:
		_, ok := x.Expr.(*syntax.Throw)
		return ok
	}
	return false
}
