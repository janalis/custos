// Package signalregistrationfailurereportedassuccess implements the native SignalRegistrationFailureReportedAsSuccess inspection.
package signalregistrationfailurereportedassuccess

import (
	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Return registration status or handle failure."

type rule struct{}

// New returns the SignalRegistrationFailureReportedAsSuccess inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalRegistrationFailureReportedAsSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "pcntl_signal") && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(n, message)
	}
}
