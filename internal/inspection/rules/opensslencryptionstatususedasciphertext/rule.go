// Package opensslencryptionstatususedasciphertext implements OpenSslEncryptionStatusUsedAsCiphertext.
package opensslencryptionstatususedasciphertext

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the encryption output as ciphertext."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslEncryptionStatusUsedAsCiphertext" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	var inputs []syntax.Expr
	switch c := n.(type) {
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "base64_encode", "bin2hex":
			inputs = []syntax.Expr{semanticquery.CallArgument(c.Args, 0, "string")}
		case "file_put_contents":
			inputs = []syntax.Expr{semanticquery.CallArgument(c.Args, 1, "data")}
		}
	case *syntax.Echo:
		inputs = c.Exprs
	}
	for _, e := range inputs {
		origin := semanticquery.OriginCall(ctx, e, "openssl_public_encrypt", "openssl_private_encrypt")
		if origin != nil {
			output, ok := semanticquery.CallArgument(origin.Args, 1, "encrypted_data").(*syntax.Variable)
			if !ok {
				continue
			}
			if status, ok := e.(*syntax.Variable); ok && status.Name == output.Name {
				continue
			}
			ctx.ReportNode(n, message)
			return
		}
	}
}
