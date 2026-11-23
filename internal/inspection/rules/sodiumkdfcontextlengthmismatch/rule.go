// Package sodiumkdfcontextlengthmismatch implements the native
// SodiumKdfContextLengthMismatch inspection.
package sodiumkdfcontextlengthmismatch

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Supply an eight-byte KDF context."

type rule struct{}

// New returns the SodiumKdfContextLengthMismatch inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumKdfContextLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP.Below(phpversion.PHP72) || !semanticquery.NativeBuiltin(ctx, c, "sodium_crypto_kdf_derive_from_key") {
		return
	}
	length, known := semanticquery.KnownByteLength(ctx, semanticquery.CallArgument(c.Args, 2, "context"))
	if known && length != 8 {
		ctx.ReportNode(c, message)
	}
}
