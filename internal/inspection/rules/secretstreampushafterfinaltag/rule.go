// Package secretstreampushafterfinaltag implements SecretstreamPushAfterFinalTag.
package secretstreampushafterfinaltag

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Start a new stream after the final tag."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SecretstreamPushAfterFinalTag" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	const name = "sodium_crypto_secretstream_xchacha20poly1305_push"
	if ctx.PHP < phpversion.PHP72 || !semanticquery.NativeBuiltin(ctx, c, name) {
		return
	}
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, p, name) || !semanticquery.NativeDominates(p, c) || !semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "state"), semanticquery.CallArgument(c.Args, 0, "state")) {
			continue
		}
		tag, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(p.Args, 3, "tag"))
		if known && tag == 3 {
			ctx.ReportNode(c, message)
			return
		}
	}
}
