// Package zipclosefailureunchecked implements the native ZipCloseFailureUnchecked inspection.
package zipclosefailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check archive finalization before returning success."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipCloseFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "close") || !semanticquery.NativeSuccessFollower(ctx, c) {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	mutationStmt, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	mutation, ok := mutationStmt.Expr.(*syntax.MethodCall)
	if !ok {
		return
	}
	if !(semanticquery.NativeMethod(ctx, mutation, "ZipArchive", "addFromString") || semanticquery.NativeMethod(ctx, mutation, "ZipArchive", "addFile")) || !astquery.Equivalent(ctx.File, mutation.Var, c.Var) {
		return
	}
	guard, ok := prior[len(prior)-2].(*syntax.If)
	if !ok || guard.Else != nil || !syntax.Terminates(guard.Body) {
		return
	}
	comparison, ok := syntax.UnwrapParens(guard.Cond).(*syntax.Binary)
	if !ok || comparison.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	open, ok := syntax.UnwrapParens(comparison.Left).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, open, "ZipArchive", "open") || !astquery.Equivalent(ctx.File, open.Var, c.Var) {
		return
	}
	truth, known := astquery.BoolConst(comparison.Right)
	if known && truth {
		ctx.ReportNode(c, message)
	}
}
