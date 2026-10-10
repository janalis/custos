// Package decryptionfailureusedasplaintext implements the native DecryptionFailureUsedAsPlaintext inspection.
package decryptionfailureusedasplaintext

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check decryption success before using the plaintext."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DecryptionFailureUsedAsPlaintext" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	sink := n.(*syntax.FuncCall)
	var data syntax.Expr
	switch semanticquery.NativeBuiltinName(ctx, sink) {
	case "strlen":
		data = semanticquery.CallArgument(sink.Args, 0, "string")
	case "file_put_contents":
		data = semanticquery.CallArgument(sink.Args, 1, "data")
	default:
		return
	}
	call, ok := semanticquery.NativeValue(ctx, data).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "openssl_decrypt") || ctx.Flow().Excludes(data, "false") || semanticquery.NativeSentinelGuard(ctx, data, "false") {
		return
	}
	ctx.ReportNode(data, message)
}
