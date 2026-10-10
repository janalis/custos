// Package transliterationfailureconsumed implements the native TransliterationFailureConsumed inspection.
package transliterationfailureconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check transliteration before consuming text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TransliterationFailureConsumed" }
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
		valid = semanticquery.NativeBuiltin(ctx, f, "transliterator_transliterate")
	}
	if m, ok := origin.(*syntax.MethodCall); ok {
		valid = semanticquery.NativeMethod(ctx, m, "Transliterator", "transliterate")
	}
	if valid && !ctx.Flow().Excludes(input, "false") && !semanticquery.NativeSentinelGuard(ctx, input, "false") {
		ctx.ReportNode(c, message)
	}
}
