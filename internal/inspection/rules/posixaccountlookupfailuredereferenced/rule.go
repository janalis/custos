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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, a.Var, n) {
		return
	}
	c, ok := semanticquery.NativeLocalValue(ctx, a.Var).(*syntax.FuncCall)
	if !ok || (!semanticquery.NativeBuiltin(ctx, c, "posix_getpwnam") && !semanticquery.NativeBuiltin(ctx, c, "posix_getpwuid")) {
		return
	}
	if !ctx.Flow().Excludes(a.Var, "false") && !semanticquery.NativeLocalSentinelGuard(ctx, a.Var, "false") {
		ctx.ReportNode(n, message)
	}
}
