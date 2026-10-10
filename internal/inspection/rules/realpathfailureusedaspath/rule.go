// Package realpathfailureusedaspath implements the native RealpathFailureUsedAsPath inspection.
package realpathfailureusedaspath

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject canonicalization failure before joining paths."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RealpathFailureUsedAsPath" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.ExpansionDPathCall(ctx, c) {
		return
	}
	b, ok := semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(c.Args, 0, "filename")).(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TDot {
		return
	}
	for _, input := range []syntax.Expr{b.Left, b.Right} {
		if !semanticquery.ExpansionDUnaliased(ctx, input, c) {
			continue
		}
		source, known := semanticquery.NativeLocalValue(ctx, input).(*syntax.FuncCall)
		if known && semanticquery.NativeBuiltin(ctx, source, "realpath") && !semanticquery.NativeLocalSentinelGuard(ctx, input, "false") && !ctx.Flow().Excludes(input, "false") {
			ctx.ReportNode(c, message)
			return
		}
	}
}
