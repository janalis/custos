// Package curlmultipartbodywithjsoncontenttype implements the native CurlMultipartBodyWithJsonContentType inspection.
package curlmultipartbodywithjsoncontenttype

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Encode the request body as JSON before declaring JSON content."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlMultipartBodyWithJsonContentType" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	body, bk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_POSTFIELDS")
	if !bk || semanticquery.NativeArray(ctx, body) == nil {
		return
	}
	headers, hk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_HTTPHEADER")
	entries, known := semanticquery.NativeArrayEntries(ctx, headers)
	if !hk || !known {
		return
	}
	for _, value := range entries {
		header, known := semanticquery.NativeString(ctx, value)
		if known && strings.EqualFold(strings.TrimSpace(header), "Content-Type: application/json") {
			ctx.ReportNode(headers, message)
			return
		}
	}
}
