// Package curlcustomgetretainspostbody implements the native CurlCustomGetRetainsPostBody inspection.
package curlcustomgetretainspostbody

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clear the request body before switching to GET."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlCustomGetRetainsPostBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	method, mk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_CUSTOMREQUEST")
	name, nk := semanticquery.NativeString(ctx, method)
	if !mk || !nk || name != "GET" {
		return
	}
	body, bk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_POSTFIELDS")
	if !bk || body.Span().Start >= method.Span().Start {
		return
	}
	for _, option := range []string{"CURLOPT_HTTPGET", "CURLOPT_POST", "CURLOPT_UPLOAD", "CURLOPT_NOBODY"} {
		if value, explicit := semanticquery.NativeCurlOption(ctx, c, h, option); explicit && value.Span().Start > body.Span().Start {
			return
		}
	}
	text, known := semanticquery.NativeString(ctx, body)
	if known && text != "" {
		ctx.ReportNode(c, message)
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, body)
	if known && len(entries) > 0 {
		ctx.ReportNode(c, message)
	}
}
