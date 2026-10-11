// Package untrustedheadervalue implements the native UntrustedHeaderValue inspection.
package untrustedheadervalue

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

const message = "Reject line breaks and validate this header value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedHeaderValue" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, call, "header") {
		value := semanticquery.CallArgument(call.Args, 0, "header")
		// The shared taint summary keeps the request source through numeric
		// results, match subjects and several line-break guards.
		if value != nil && ctx.Flow().Tainted(value, flow.Header) && inert(ctx, value, maxDepth) {
			return
		}
	}
	semanticquery.CheckUntrustedHeaderValue(ctx, n, message)
}
