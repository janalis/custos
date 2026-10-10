// Package copyfailurereportedassuccess implements the native CopyFailureReportedAsSuccess inspection.
package copyfailurereportedassuccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the copy result or handle failure."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CopyFailureReportedAsSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "copy") {
		return
	}
	if syntax.EnclosingVariableScope(c) != nil && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(c, message)
	}
}
