// Package opensslrawciphertextoptionmismatch implements the native OpenSslRawCiphertextOptionMismatch inspection.
package opensslrawciphertextoptionmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match raw-ciphertext options when decrypting."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslRawCiphertextOptionMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_decrypt") {
		return
	}
	data := semanticquery.CallArgument(call.Args, 0, "data")
	encrypt, ok := semanticquery.NativeValue(ctx, data).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, encrypt, "openssl_encrypt") {
		return
	}
	raw, known := semanticquery.NativeFlag(ctx, semanticquery.CallArgument(call.Args, 3, "options"), 1)
	encoded, ek := semanticquery.NativeFlag(ctx, semanticquery.CallArgument(encrypt.Args, 3, "options"), 1)
	if !known || !ek || raw == encoded {
		return
	}
	for _, parameter := range []struct {
		pos  int
		name string
	}{{1, "cipher_algo"}, {2, "passphrase"}, {4, "iv"}} {
		a := semanticquery.CallArgument(call.Args, parameter.pos, parameter.name)
		b := semanticquery.CallArgument(encrypt.Args, parameter.pos, parameter.name)
		if a == nil || b == nil || !astquery.Equivalent(ctx.File, a, b) {
			return
		}
		av, bv := ctx.Flow().Value(a), ctx.Flow().Value(b)
		if !av.Complete || !bv.Complete || av.Invalidated != bv.Invalidated {
			return
		}
		if _, ok := a.(*syntax.Variable); ok && (av.Expr == nil || av.Expr != bv.Expr) {
			return
		}
	}
	ctx.ReportNode(call, message)
}
