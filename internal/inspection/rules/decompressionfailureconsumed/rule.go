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
	var origin *syntax.FuncCall
	for _, candidate := range semanticquery.NativeLocalBuiltinResults(ctx, input, "false", "gzdecode", "gzuncompress", "gzinflate", "inflate_add") {
		if call, ok := candidate.(*syntax.FuncCall); ok && !safeCompression(ctx, call) {
			origin = call
			break
		}
	}
	ok := origin != nil
	if !ok {
		return
	}

	// Every strict false guard on the tracked local is a fact of the shared
	// flow analysis, which also provides the reaching decompression call.
	if !ctx.Flow().Excludes(input, "false") {
		ctx.ReportNode(c, message)
	}
}

// safeCompression recognizes a statically valid matching encoder. Exclude
// these producers before selection so they cannot hide another failing call.
func safeCompression(ctx *analysis.Context, origin *syntax.FuncCall) bool {
	encoder, known := semanticquery.NativeValue(ctx, semanticquery.CallArgument(origin.Args, 0, "data")).(*syntax.FuncCall)
	if known && semanticquery.CallArgument(encoder.Args, 2, "encoding") == nil {
		expected := map[string]string{"gzdecode": "gzencode", "gzuncompress": "gzcompress", "gzinflate": "gzdeflate"}[semanticquery.NativeBuiltinName(ctx, origin)]
		if expected != "" && semanticquery.NativeBuiltin(ctx, encoder, expected) {
			level := semanticquery.CallArgument(encoder.Args, 1, "level")
			if level == nil {
				return true
			}
			if n, ok := semanticquery.NativeInt(ctx, level); ok && n >= -1 && n <= 9 {
				return true
			}
		}
	}

	return false
}
