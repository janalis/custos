// Package compressiondecoderformatmismatch implements the native CompressionDecoderFormatMismatch inspection.
package compressiondecoderformatmismatch

import (
	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Match the decoder to the compression format."

type rule struct{}

// New returns the CompressionDecoderFormatMismatch inspection.
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
	if !semanticquery.StableLocalValue(ctx, arg, semanticquery.LocalStability{CopiesOnly: true}) || !defaultFormat(encoder) {
		return
	}

	match := map[string]string{"gzencode": "gzdecode", "gzcompress": "gzuncompress", "gzdeflate": "gzinflate"}[semanticquery.NativeBuiltinName(ctx, encoder)]
	if match != "" && match != decoder {
		var fixes []diagnostic.Fix
		if fn := ctx.Index().Function(match, ctx.PHP); fn != nil && fn.Builtin && (match != "gzdecode" || ctx.PHP.AtLeast(phpversion.PHP54)) {
			fixes = append(fixes, astquery.ReplaceFix(c.Name.Span(), semanticquery.QualifiedBuiltinFor(ctx, match, c)))
		}
		ctx.ReportNode(c, message, fixes...)
	}
}

// defaultFormat reports whether the encoder call provably omits the encoding
// argument: at most the data and level slots are bound, and nothing is unpacked.
func defaultFormat(encoder *syntax.FuncCall) bool {
	for _, node := range encoder.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || (arg.Name != nil && arg.Name.Value == "encoding") {
			return false
		}
	}
	return semanticquery.CallArgument(encoder.Args, 2, "encoding") == nil
}
