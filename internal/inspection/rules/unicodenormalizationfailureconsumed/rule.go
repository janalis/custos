// Package unicodenormalizationfailureconsumed implements the native UnicodeNormalizationFailureConsumed inspection.
package unicodenormalizationfailureconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check normalization before consuming text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnicodeNormalizationFailureConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "strlen", "strtoupper":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "string")
	origin := semanticquery.NativeValue(ctx, input)
	valid := false
	if f, ok := origin.(*syntax.FuncCall); ok {
		valid = semanticquery.NativeBuiltin(ctx, f, "normalizer_normalize")
	}
	valid = valid || semanticquery.ExpansionCStatic(ctx, origin, "Normalizer", "normalize")
	if valid && !ctx.Flow().Excludes(input, "false") && !semanticquery.NativeSentinelGuard(ctx, input, "false") {
		ctx.ReportNode(c, message)
	}
}
