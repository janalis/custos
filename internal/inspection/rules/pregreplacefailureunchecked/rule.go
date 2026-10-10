// Package pregreplacefailureunchecked implements the native PregReplaceFailureUnchecked inspection.
package pregreplacefailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle regex replacement failure before using its result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PregReplaceFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	for _, input := range semanticquery.NativeStringInputs(ctx, call, call.Args) {
		inner, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
		if !ok {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, inner) {
		case "preg_replace", "preg_replace_callback":
		default:
			continue
		}
		if ctx.TypeOf(input).OnlyOf("string") || ctx.Flow().Excludes(input, "null") || semanticquery.NativeSentinelGuard(ctx, input, "null") {
			continue
		}
		ctx.ReportNode(call, message)
		return
	}
}
