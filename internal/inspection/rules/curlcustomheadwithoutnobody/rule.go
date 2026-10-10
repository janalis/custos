// Package curlcustomheadwithoutnobody implements the native CurlCustomHeadWithoutNoBody inspection.
package curlcustomheadwithoutnobody

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use CURLOPT_NOBODY for a HEAD transfer."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlCustomHeadWithoutNoBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	method, known := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_CUSTOMREQUEST")
	name, nk := semanticquery.NativeString(ctx, method)
	if !known || !nk || name != "HEAD" {
		return
	}
	noBody, explicit := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_NOBODY")
	truth, tk := semanticquery.NativeTruth(ctx, noBody)
	if get, known := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_HTTPGET"); known {
		value, known := semanticquery.NativeTruth(ctx, get)
		if !known {
			return
		}
		if value && (!explicit || get.Span().Start > noBody.Span().Start) {
			ctx.ReportNode(c, message)
			return
		}
	}
	if explicit && tk && !truth || !explicit && semanticquery.NativeCurlDefault(ctx, c, h, "CURLOPT_NOBODY") {
		ctx.ReportNode(c, message)
	}
}
