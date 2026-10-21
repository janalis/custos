// Package detachedsignaturepassedtocombinedverifier implements DetachedSignaturePassedToCombinedVerifier.
package detachedsignaturepassedtocombinedverifier

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Verify detached signatures with the detached API."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DetachedSignaturePassedToCombinedVerifier" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP >= phpversion.PHP72 && semanticquery.NativeBuiltin(ctx, c, "sodium_crypto_sign_open") && semanticquery.OriginCall(ctx, semanticquery.CallArgument(c.Args, 0, "signed_message"), "sodium_crypto_sign_detached") != nil {
		ctx.ReportNode(c, message)
	}
}
