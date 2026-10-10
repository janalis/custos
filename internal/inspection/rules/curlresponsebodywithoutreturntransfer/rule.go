// Package curlresponsebodywithoutreturntransfer implements the native CurlResponseBodyWithoutReturnTransfer inspection.
package curlresponsebodywithoutreturntransfer

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable response-body return before decoding curl output."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlResponseBodyWithoutReturnTransfer" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	for _, input := range semanticquery.NativeStringInputs(ctx, call, call.Args) {
		producer, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, producer, "curl_exec") {
			continue
		}
		handle := semanticquery.CallArgument(producer.Args, 0, "handle")
		origin, known := ctx.Flow().Resolve(handle)
		if !known {
			continue
		}
		init, ok := origin.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, init, "curl_init") {
			continue
		}
		if !semanticquery.NativeCurlDefault(ctx, producer, handle, "CURLOPT_RETURNTRANSFER") {
			continue
		}
		ctx.ReportNode(call, message)
		return
	}
}
