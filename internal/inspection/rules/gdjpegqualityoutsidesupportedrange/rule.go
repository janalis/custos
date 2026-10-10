// Package gdjpegqualityoutsidesupportedrange implements the native GdJpegQualityOutsideSupportedRange inspection.
package gdjpegqualityoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply JPEG quality between zero and 100."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdJpegQualityOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagejpeg") {
		return
	}
	v, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 2, "quality"))
	if known && (v < -1 || v > 100) {
		ctx.ReportNode(n, message)
	}
}
