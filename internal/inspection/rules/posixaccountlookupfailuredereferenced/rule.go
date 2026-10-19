// Package posixaccountlookupfailuredereferenced implements the native PosixAccountLookupFailureDereferenced inspection.
package posixaccountlookupfailuredereferenced

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject account lookup failure before indexing."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PosixAccountLookupFailureDereferenced" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, a.Var, n) && !semanticquery.NativeFailureFlowOnly(ctx, a.Var) {
		return
	}
	// Among several reaching assignments only a lookup result counts: an
	// unguarded result of another call does not stand in for it.
	c, ok := semanticquery.NativeLocalBuiltinResult(ctx, a.Var, "false", "posix_getpwnam", "posix_getpwuid").(*syntax.FuncCall)
	if !ok {
		return
	}
	if ctx.Flow().Excludes(a.Var, "false") || !semanticquery.NativeFailureFlowOnly(ctx, a.Var) && semanticquery.NativeLocalSentinelGuard(ctx, a.Var, "false") {
		return
	}
	if semanticquery.NativeFailureFlowOnly(ctx, a.Var) || !(guard{ctx: ctx, target: a.Var, call: c}).covers() {
		ctx.ReportNode(n, message)
	}
}
