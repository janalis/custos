// Package nestedpdotransactionwithoutsavepoint implements the native NestedPdoTransactionWithoutSavepoint inspection.
package nestedpdotransactionwithoutsavepoint

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Finish the transaction before beginning another one."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NestedPdoTransactionWithoutSavepoint" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "beginTransaction") {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) == 0 {
		return
	}
	statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	previous, ok := statement.Expr.(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, previous, "PDO", "beginTransaction") && semanticquery.NativeValue(ctx, c.Var) != nil && semanticquery.NativeValue(ctx, c.Var) == semanticquery.NativeValue(ctx, previous.Var) {
		ctx.ReportNode(c, message)
	}
}
