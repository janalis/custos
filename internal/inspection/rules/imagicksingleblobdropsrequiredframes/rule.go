// Package imagicksingleblobdropsrequiredframes implements the native ImagickSingleBlobDropsRequiredFrames inspection.
package imagicksingleblobdropsrequiredframes

import (
	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Export every frame when frame preservation is enabled."

type rule struct{}

// New returns the ImagickSingleBlobDropsRequiredFrames inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImagickSingleBlobDropsRequiredFrames" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if semanticquery.NativeMethod(ctx, c, "Imagick", "getImageBlob") && ctx.Bool("preserveFrames") && semanticquery.AppendedFrames(ctx, c) >= 2 {
		ctx.ReportNode(n, message)
	}
}
