// Package decompressionfailureconsumed implements the native DecompressionFailureConsumed inspection.
package decompressionfailureconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check decompression before consuming data."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DecompressionFailureConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "strlen", "strtoupper":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "string")
	origin, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok {
		return
	}
	switch semanticquery.NativeBuiltinName(ctx, origin) {
	case "gzdecode", "gzuncompress", "gzinflate", "inflate_add":
	default:
		return
	}
	encoder, known := semanticquery.NativeValue(ctx, semanticquery.CallArgument(origin.Args, 0, "data")).(*syntax.FuncCall)
	if known && semanticquery.CallArgument(encoder.Args, 2, "encoding") == nil {
		expected := map[string]string{"gzdecode": "gzencode", "gzuncompress": "gzcompress", "gzinflate": "gzdeflate"}[semanticquery.NativeBuiltinName(ctx, origin)]
		if expected != "" && semanticquery.NativeBuiltin(ctx, encoder, expected) {
			level := semanticquery.CallArgument(encoder.Args, 1, "level")
			if level == nil {
				return
			}
			if n, ok := semanticquery.NativeInt(ctx, level); ok && n >= -1 && n <= 9 {
				return
			}
		}
	}

	if !ctx.Flow().Excludes(input, "false") && !semanticquery.NativeSentinelGuard(ctx, input, "false") {
		ctx.ReportNode(c, message)
	}
}
