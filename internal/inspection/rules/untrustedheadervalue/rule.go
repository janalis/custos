// Package untrustedheadervalue implements the native UntrustedHeaderValue inspection.
package untrustedheadervalue

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/flow"
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
	// A partial application of header() is a sink too, so the name is
	// matched rather than a call yielding header()'s result.
	if semanticquery.NativeBuiltinName(ctx, call) == "header" {
		value := semanticquery.CallArgument(call.Args, 0, "header")
		// The shared taint summary keeps the request source through numeric
		// results, match subjects and several line-break guards.
		if value != nil && ctx.Flow().Tainted(value, flow.Header) && inert(ctx, value, maxDepth) {
			return
		}
	}
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name == "header" && ctx.Flow().Tainted(semanticquery.CallArgument(c.Args, 0, "header"), flow.Header) || semanticquery.TaintedWrapperSink(ctx, c, flow.Header) {
		ctx.ReportNode(c, message)
	}
}

// PartialCalls: a partial application binds the arguments it is written
// with, and every call of the closure passes them to the sink.
func (rule) PartialCalls() {}
