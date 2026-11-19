// Package imagedecodefailureunchecked implements the native ImageDecodeFailureUnchecked inspection.
package imagedecodefailureunchecked

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

const message = "Check image decoding success before using the image."

type rule struct{}

// New returns the ImageDecodeFailureUnchecked inspection.
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
	if semanticquery.NativeFailureExcluded(ctx, input, types.Of(`\GdImage`, "false"), "false") {
		return
	}
	ctx.ReportNode(input, message)
}
