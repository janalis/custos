// Package gdimagedimensionsnotpositive implements the native GdImageDimensionsNotPositive inspection.
package gdimagedimensionsnotpositive

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply positive image dimensions."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdImageDimensionsNotPositive" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagecreate") && !semanticquery.NativeBuiltin(ctx, c, "imagecreatetruecolor") {
		return
	}
	w, wk := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "width"))
	h, hk := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 1, "height"))
	if (wk && w <= 0) || (hk && h <= 0) {
		ctx.ReportNode(n, message)
	}
}
