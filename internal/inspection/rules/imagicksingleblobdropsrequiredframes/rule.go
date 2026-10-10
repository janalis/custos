// Package imagicksingleblobdropsrequiredframes implements the native ImagickSingleBlobDropsRequiredFrames inspection.
package imagicksingleblobdropsrequiredframes

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Export every frame when frame preservation is enabled."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImagickSingleBlobDropsRequiredFrames" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if semanticquery.NativeMethod(ctx, c, "Imagick", "getImageBlob") && ctx.Bool("preserveFrames") && semanticquery.ExpansionDFrames(ctx, c) >= 2 {
		ctx.ReportNode(n, message)
	}
}
