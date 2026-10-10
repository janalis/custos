// Package curlexplicitinfinitetimeout implements the native CurlExplicitInfiniteTimeout inspection.
package curlexplicitinfinitetimeout

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Set a finite transfer timeout."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlExplicitInfiniteTimeout" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	seconds, sk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_TIMEOUT")
	millis, mk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_TIMEOUT_MS")
	s, sn := semanticquery.NativeInt(ctx, seconds)
	ms, mn := semanticquery.NativeInt(ctx, millis)
	if sk && (!sn || s > 0) || mk && (!mn || ms > 0) {
		return
	}
	if sk && s == 0 || mk && ms == 0 {
		ctx.ReportNode(c, message)
	}
}
