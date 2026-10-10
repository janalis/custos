// Package intlformattingfailureunchecked implements the native IntlFormattingFailureUnchecked inspection.
package intlformattingfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check formatting success before consuming the formatted text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlFormattingFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "strtoupper", "strlen":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "string")
	origin, ok := semanticquery.NativeValue(ctx, input).(*syntax.MethodCall)
	if !ok || !(semanticquery.NativeMethod(ctx, origin, "NumberFormatter", "format") || semanticquery.NativeMethod(ctx, origin, "IntlDateFormatter", "format")) {
		return
	}
	if ctx.Flow().Excludes(input, "false") || semanticquery.NativeSentinelGuard(ctx, input, "false") {
		return
	}
	ctx.ReportNode(input, message)
}
