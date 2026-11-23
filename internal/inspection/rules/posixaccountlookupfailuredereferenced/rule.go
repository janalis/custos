// Package posixaccountlookupfailuredereferenced implements the native PosixAccountLookupFailureDereferenced inspection.
package posixaccountlookupfailuredereferenced

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

const message = "Reject account lookup failure before indexing."

type rule struct{}

// New returns the PosixAccountLookupFailureDereferenced inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PosixAccountLookupFailureDereferenced" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.MembersUnwrittenBefore(ctx, a.Var, n) && !semanticquery.NativeFailureFlowOnly(ctx, a.Var) {
		return
	}
	// Among several reaching assignments only a lookup result counts: an
	// unguarded result of another call does not stand in for it.
	if _, ok := semanticquery.NativeLocalBuiltinResult(ctx, a.Var, "false", "posix_getpwnam", "posix_getpwuid").(*syntax.FuncCall); !ok {
		return
	}
	if semanticquery.NativeFailureExcludedSingle(ctx, a.Var, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(n, message)
}
