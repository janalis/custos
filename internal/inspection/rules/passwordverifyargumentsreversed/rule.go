// Package passwordverifyargumentsreversed implements PasswordVerifyArgumentsReversed.
package passwordverifyargumentsreversed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Pass the password before its stored hash."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordVerifyArgumentsReversed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP55 || !semanticquery.NativeBuiltin(ctx, c, "password_verify") {
		return
	}
	origin := semanticquery.OriginCall(ctx, semanticquery.CallArgument(c.Args, 0, "password"), "password_hash")
	if origin != nil && semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(origin.Args, 0, "password"), semanticquery.CallArgument(c.Args, 1, "hash")) {
		ctx.ReportNode(c, message)
	}
}
