// Package imagedecodefailureunchecked implements the native ImageDecodeFailureUnchecked inspection.
package imagedecodefailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check image decoding success before using the image."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImageDecodeFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "imagepng", "imagejpeg", "imagegif", "imagesx", "imagesy":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "image")
	origin, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, origin, "imagecreatefromstring") {
		return
	}
	if _, literal := semanticquery.NativeString(ctx, semanticquery.CallArgument(origin.Args, 0, "data")); literal {
		return
	}
	if ctx.Flow().Excludes(input, "false") || semanticquery.NativeSentinelGuard(ctx, input, "false") {
		return
	}
	ctx.ReportNode(input, message)
}
