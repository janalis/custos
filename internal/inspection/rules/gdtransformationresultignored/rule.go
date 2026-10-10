// Package gdtransformationresultignored implements the native GdTransformationResultIgnored inspection.
package gdtransformationresultignored

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use the image returned by the transformation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdTransformationResultIgnored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "imagepng", "imagejpeg", "imagegif", "imagewebp", "imageavif":
	default:
		return
	}
	e := semanticquery.CallArgument(c.Args, 0, "image")
	if !semanticquery.ExpansionDUnaliased(ctx, e, n) {
		return
	}
	for _, p := range semanticquery.NativeStreamCalls(ctx, c, e, "imagescale", "imagecrop") {
		if _, discarded := p.Parent().(*syntax.ExprStmt); discarded {
			ctx.ReportNode(n, message)
			return
		}
	}
}
