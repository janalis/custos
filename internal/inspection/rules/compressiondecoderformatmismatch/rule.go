// Package compressiondecoderformatmismatch implements the native CompressionDecoderFormatMismatch inspection.
package compressiondecoderformatmismatch

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Match the decoder to the compression format."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CompressionDecoderFormatMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	decoder := semanticquery.NativeBuiltinName(ctx, c)
	switch decoder {
	case "gzdecode", "gzuncompress", "gzinflate":
	default:
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "data")
	encoder, ok := semanticquery.NativeValue(ctx, arg).(*syntax.FuncCall)
	if !ok {
		return
	}
	if semanticquery.CallArgument(encoder.Args, 2, "encoding") != nil {
		return
	}

	match := map[string]string{"gzencode": "gzdecode", "gzcompress": "gzuncompress", "gzdeflate": "gzinflate"}[semanticquery.NativeBuiltinName(ctx, encoder)]
	if match != "" && match != decoder {
		var fixes []diagnostic.Fix
		if fn := ctx.Index().Function(match, ctx.PHP); fn != nil && fn.Builtin && (match != "gzdecode" || ctx.PHP >= phpversion.PHP54) {
			fixes = append(fixes, astquery.ReplaceFix(c.Name.Span(), semanticquery.QualifiedBuiltinFor(ctx, match, c)))
		}
		ctx.ReportNode(c, message, fixes...)
	}
}
