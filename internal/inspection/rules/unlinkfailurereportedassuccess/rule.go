// Package unlinkfailurereportedassuccess implements the native UnlinkFailureReportedAsSuccess inspection.
package unlinkfailurereportedassuccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the deletion result or handle failure."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnlinkFailureReportedAsSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "unlink") {
		return
	}
	if syntax.EnclosingVariableScope(c) != nil && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(c, message)
	}
}
