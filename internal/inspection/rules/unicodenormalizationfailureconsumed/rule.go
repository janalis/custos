// Package unicodenormalizationfailureconsumed implements the native UnicodeNormalizationFailureConsumed inspection.
package unicodenormalizationfailureconsumed

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

const message = "Check normalization before consuming text."

type rule struct{}

// New returns the UnicodeNormalizationFailureConsumed inspection.
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
	origin := semanticquery.NativeLocalBuiltinResult(ctx, input, "false", "normalizer_normalize", "Normalizer::normalize")
	valid := false
	if f, ok := origin.(*syntax.FuncCall); ok {
		valid = semanticquery.NativeBuiltin(ctx, f, "normalizer_normalize")
	}
	valid = valid || semanticquery.BuiltinStaticCall(ctx, origin, "Normalizer", "normalize")
	if valid && !ctx.Flow().Excludes(input, "false") && (semanticquery.NativeFailureFlowOnly(ctx, input) || !semanticquery.NativeNarrowedExcludes(ctx, input, types.Of("string", "false"), "false")) {
		ctx.ReportNode(c, message)
	}
}
