// Package jsoninvaliddepth implements the native JsonInvalidDepth inspection.
package jsoninvaliddepth

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a supported JSON depth."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonInvalidDepth" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "json_decode" && name != "json_encode" {
		return
	}
	d, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 2, "depth"))
	if k && (d <= 0 || name == "json_decode" && d > 2147483647) {
		ctx.ReportNode(c, message)
	}
}
