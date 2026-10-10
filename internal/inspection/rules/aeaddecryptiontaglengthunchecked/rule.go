// Package aeaddecryptiontaglengthunchecked implements the native AeadDecryptionTagLengthUnchecked inspection.
package aeaddecryptiontaglengthunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Validate the authentication tag length before decrypting."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AeadDecryptionTagLengthUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_decrypt") {
		return
	}
	cipher, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "cipher_algo"))
	if !known {
		return
	}
	switch strings.ToLower(cipher) {
	case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm":
	default:
		return
	}
	tag := semanticquery.CallArgument(call.Args, 5, "tag")
	if _, ok := request(ctx, tag); !ok {
		return
	}
	size, known := ctx.Flow().ExactLength(tag)
	if (!known || size != int64(ctx.Int("tagLength"))) && !semanticquery.NativeLengthGuard(ctx, tag, int64(ctx.Int("tagLength"))) {
		ctx.ReportNode(tag, message)
	}
}

func request(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	x, ok := semanticquery.NativeValue(ctx, e).(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	v, ok := x.Var.(*syntax.Variable)
	if !ok {
		return "", false
	}
	switch v.Name {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_SERVER":
	default:
		return "", false
	}
	if !semanticquery.NativeRequestUnwritten(ctx, x) {
		return "", false
	}
	key, known := semanticquery.NativeString(ctx, x.Dim)
	return key, known
}
