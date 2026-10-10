// Package curlmethodoptionorderconflict implements the native CurlMethodOptionOrderConflict inspection.
package curlmethodoptionorderconflict

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Configure a consistent final HTTP request method."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlMethodOptionOrderConflict" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	post, pk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_POST")
	get, gk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_HTTPGET")
	pv, pn := semanticquery.NativeTruth(ctx, post)
	gv, gn := semanticquery.NativeTruth(ctx, get)
	if !pk || !gk || !pn || !gn || !pv || !gv {
		return
	}
	later := post
	if get.Span().Start > post.Span().Start {
		later = get
	}
	call := later.Parent().Parent().Parent().(*syntax.FuncCall)
	ctx.ReportNode(semanticquery.CallArgument(call.Args, 1, "option"), message)
}
