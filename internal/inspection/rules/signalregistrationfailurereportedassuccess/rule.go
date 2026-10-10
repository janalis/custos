// Package signalregistrationfailurereportedassuccess implements the native SignalRegistrationFailureReportedAsSuccess inspection.
package signalregistrationfailurereportedassuccess

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return registration status or handle failure."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalRegistrationFailureReportedAsSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "pcntl_signal") && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(n, message)
	}
}
