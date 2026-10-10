// Package sodiumnoncelengthmismatch implements the native SodiumNonceLengthMismatch inspection.
package sodiumnoncelengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use the required secretbox nonce length."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumNonceLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP72 || !semanticquery.NativeBuiltin(ctx, c, "sodium_crypto_secretbox") {
		return
	}
	nonce := semanticquery.CallArgument(c.Args, 1, "nonce")
	length, known := byteLength(ctx, nonce)
	if known && length != 24 {
		if c, ok := nonce.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, c, "random_bytes") {
			size := semanticquery.CallArgument(c.Args, 0, "length")
			if literal, ok := size.(*syntax.Literal); ok && literal.LitKind == syntax.LitInt {
				ctx.ReportNode(nonce, message, astquery.ReplaceFix(size.Span(), "\\SODIUM_CRYPTO_SECRETBOX_NONCEBYTES"))
				return
			}
		}
		ctx.ReportNode(nonce, message)
	}
}

func byteLength(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if text, ok := semanticquery.NativeString(ctx, e); ok {
		return int64(len(text)), true
	}
	c, ok := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return 0, false
	}
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "random_bytes":
		size, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 0, "length"))
		return size, ok && size > 0
	case "str_repeat":
		text, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "string"))
		count, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "times"))
		if !ok || !known || count < 0 || count > 1<<30 {
			return 0, false
		}
		return int64(len(text)) * count, true
	}
	return 0, false
}
