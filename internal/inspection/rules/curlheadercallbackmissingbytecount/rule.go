// Package curlheadercallbackmissingbytecount implements CurlHeaderCallbackMissingByteCount.
package curlheadercallbackmissingbytecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the consumed header byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderCallbackMissingByteCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := semanticquery.CurlSetting(ctx, c, "CURLOPT_HEADERFUNCTION")
	if value == nil {
		return
	}
	_, body := semanticquery.NativeCallback(ctx, value)
	block, ok := body.(*syntax.Block)
	if !ok || semanticquery.NativeGeneratorBody(block) {
		return
	}
	missing := !semanticquery.ExpansionDCallbackTerminates(ctx, block)
	syntax.Inspect(block, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if r, ok := n.(*syntax.Return); ok && r.Expr == nil && semanticquery.NativeCallbackReachable(ctx, r) {
			missing = true
		}
		return true
	})
	if missing {
		ctx.ReportNode(c, message)
	}
}
