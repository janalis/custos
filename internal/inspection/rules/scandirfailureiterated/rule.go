// Package scandirfailureiterated implements the native ScandirFailureIterated inspection.
package scandirfailureiterated

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
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
	if ok && semanticquery.NativeBuiltin(ctx, c, "scandir") && !semanticquery.NativeFailureExcludedSingle(ctx, loop.Expr, types.Of("array", "false"), "false") {
		ctx.ReportNode(loop.Expr, message)
	}
}
