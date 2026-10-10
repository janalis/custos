// Package curlheadersincludedindecodedbody implements the native CurlHeadersIncludedInDecodedBody inspection.
package curlheadersincludedindecodedbody

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep HTTP headers separate from the decoded JSON body."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeadersIncludedInDecodedBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_decode") {
		return
	}
	producer, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "json")).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, producer, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(producer.Args, 0, "handle")
	header, hk := semanticquery.NativeCurlOption(ctx, producer, h, "CURLOPT_HEADER")
	transfer, tk := semanticquery.NativeCurlOption(ctx, producer, h, "CURLOPT_RETURNTRANSFER")
	hv, hknown := semanticquery.NativeTruth(ctx, header)
	tv, tknown := semanticquery.NativeTruth(ctx, transfer)
	if hk && tk && hknown && tknown && hv && tv {
		ctx.ReportNode(c, message)
	}
}
