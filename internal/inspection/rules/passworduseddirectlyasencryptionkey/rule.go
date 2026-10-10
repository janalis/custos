// Package passworduseddirectlyasencryptionkey implements the native PasswordUsedDirectlyAsEncryptionKey inspection.
package passworduseddirectlyasencryptionkey

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Derive an encryption key from the password."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordUsedDirectlyAsEncryptionKey" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "openssl_encrypt" && name != "openssl_decrypt" {
		return
	}
	key := semanticquery.CallArgument(call.Args, 2, "passphrase")
	field, ok := request(ctx, key)
	if !ok {
		return
	}
	for _, name := range ctx.List("passwordKeys") {
		if field == name {
			ctx.ReportNode(key, message)
			return
		}
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
