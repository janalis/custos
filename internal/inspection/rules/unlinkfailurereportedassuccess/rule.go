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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "unlink") {
		return
	}
	if syntax.EnclosingVariableScope(c) != nil && !scratch(ctx, c) && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

// scratch reports an unlink of the file tempnam() made in the function
// (E3): deleting it cleans up after the function's own work, whose success
// the return reports.
func scratch(ctx *analysis.Context, c *syntax.FuncCall) bool {
	path := semanticquery.CallArgument(c.Args, 0, "filename")
	if path == nil {
		return false
	}
	made, ok := semanticquery.NativeLocalValue(ctx, path).(*syntax.FuncCall)
	return ok && semanticquery.NativeBuiltin(ctx, made, "tempnam")
}
