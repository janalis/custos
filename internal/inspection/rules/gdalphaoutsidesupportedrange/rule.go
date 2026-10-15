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
	alpha := semanticquery.CallArgument(c.Args, 4, "alpha")
	lo, hi, known := semanticquery.NativeIntRange(ctx, alpha)
	if known && (hi < 0 || lo > 127) {
		ctx.ReportNode(n, message)
		return
	}
	if first, ok := semanticquery.NativeIntEntry(ctx, alpha); ok && (first < 0 || first > 127) {
		ctx.ReportNode(n, message)
	}
}
