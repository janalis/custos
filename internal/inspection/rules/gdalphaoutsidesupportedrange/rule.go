// Package gdalphaoutsidesupportedrange implements the native GdAlphaOutsideSupportedRange inspection.
package gdalphaoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply GD alpha between zero and 127."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdAlphaOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagecolorallocatealpha") && !semanticquery.NativeBuiltin(ctx, c, "imagecolorclosestalpha") {
		return
	}
	v, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 4, "alpha"))
	if known && (v < 0 || v > 127) {
		ctx.ReportNode(n, message)
	}
}
