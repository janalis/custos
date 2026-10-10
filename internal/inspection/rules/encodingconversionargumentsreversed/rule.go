// Package encodingconversionargumentsreversed implements the native EncodingConversionArgumentsReversed inspection.
package encodingconversionargumentsreversed

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Convert to the encoding required by the output consumer."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EncodingConversionArgumentsReversed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_encode") {
		return
	}
	conversion, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "value")).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, conversion, "mb_convert_encoding") {
		return
	}
	source, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(conversion.Args, 0, "string"))
	if !known || !utf8.ValidString(source) {
		return
	}
	from, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(conversion.Args, 2, "from_encoding"))
	if !known || !strings.EqualFold(from, "UTF-8") {
		return
	}
	to, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(conversion.Args, 1, "to_encoding"))
	if !known {
		return
	}
	switch strings.ToUpper(to) {
	case "ISO-8859-1", "ISO8859-1", "LATIN1":
	default:
		return
	}
	nonascii := false
	converted := make([]byte, 0, len(source))
	for _, ch := range source {
		if ch > 255 {
			return
		}
		converted = append(converted, byte(ch))
		if ch >= 128 {
			nonascii = true
		}
	}
	if nonascii && !utf8.Valid(converted) {
		ctx.ReportNode(conversion, message)
	}
}
