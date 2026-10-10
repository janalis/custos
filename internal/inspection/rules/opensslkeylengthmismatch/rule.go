// Package opensslkeylengthmismatch implements the native OpenSslKeyLengthMismatch inspection.
package opensslkeylengthmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Provide the cipher-required key length."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslKeyLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "openssl_encrypt" && name != "openssl_decrypt" {
		return
	}
	cipher, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "cipher_algo"))
	if !known {
		return
	}
	parts := strings.Split(strings.ToLower(cipher), "-")
	if len(parts) != 3 || parts[0] != "aes" {
		return
	}
	expected := int64(0)
	switch parts[1] {
	case "128":
		expected = 16
	case "192":
		expected = 24
	case "256":
		expected = 32
	default:
		return
	}
	switch parts[2] {
	case "cbc", "gcm", "ccm", "ctr", "ecb", "cfb", "cfb1", "cfb8", "ofb":
	default:
		return
	}
	key := semanticquery.CallArgument(call.Args, 2, "passphrase")
	size, ok := length(ctx, key)
	if ok && size != expected {
		ctx.ReportNode(key, message)
	}
}

func length(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if s, ok := semanticquery.NativeString(ctx, e); ok {
		return int64(len(s)), true
	}
	call, ok := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return 0, false
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "random_bytes":
		n, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 0, "length"))
		return n, ok && n > 0
	case "str_repeat":
		s, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "string"))
		n, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "times"))
		if ok && known && n >= 0 && n <= 1048576 {
			return int64(len(s)) * n, true
		}
	}
	return 0, false
}
