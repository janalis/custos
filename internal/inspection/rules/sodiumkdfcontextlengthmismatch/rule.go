// Package sodiumkdfcontextlengthmismatch implements SodiumKdfContextLengthMismatch.
package sodiumkdfcontextlengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Supply an eight-byte KDF context."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumKdfContextLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP72 || !semanticquery.NativeBuiltin(ctx, c, "sodium_crypto_kdf_derive_from_key") {
		return
	}
	length, known := semanticquery.KnownByteLength(ctx, semanticquery.CallArgument(c.Args, 2, "context"))
	if known && length != 8 {
		ctx.ReportNode(c, message)
	}
}
