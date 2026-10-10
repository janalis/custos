// Package incrementalinflateencodingmismatch implements the native IncrementalInflateEncodingMismatch inspection.
package incrementalinflateencodingmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the inflate encoding to the input."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IncrementalInflateEncodingMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "inflate_add") {
		return
	}
	init, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "context")).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, init, "inflate_init") {
		return
	}
	encoder, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "data")).(*syntax.FuncCall)
	if !ok {
		return
	}
	if semanticquery.CallArgument(encoder.Args, 2, "encoding") != nil {
		return
	}

	wanted := map[string]string{"gzencode": "ZLIB_ENCODING_GZIP", "gzcompress": "ZLIB_ENCODING_DEFLATE", "gzdeflate": "ZLIB_ENCODING_RAW"}[semanticquery.NativeBuiltinName(ctx, encoder)]
	if wanted == "" {
		return
	}
	encoding := semanticquery.CallArgument(init.Args, 0, "encoding")
	for _, name := range []string{"ZLIB_ENCODING_GZIP", "ZLIB_ENCODING_DEFLATE", "ZLIB_ENCODING_RAW"} {
		if name != wanted && hasConstant(ctx, encoding, name) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func hasConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	value := semanticquery.NativeValue(ctx, e)
	c, ok := value.(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(c.Name.Value, name) && semanticquery.BareReachesGlobal(ctx, name, c.Span().Start)
}
