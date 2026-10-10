// Package gdpngcompressionoutsidesupportedrange implements the native GdPngCompressionOutsideSupportedRange inspection.
package gdpngcompressionoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply PNG compression between zero and nine."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdPngCompressionOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagepng") {
		return
	}
	v, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 2, "quality"))
	if known && (v < -1 || v > 9) {
		ctx.ReportNode(n, message)
	}
}
