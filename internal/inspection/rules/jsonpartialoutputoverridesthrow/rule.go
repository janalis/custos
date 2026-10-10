// Package jsonpartialoutputoverridesthrow implements the native JsonPartialOutputOverridesThrow inspection.
package jsonpartialoutputoverridesthrow

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Choose either partial JSON output or throwing on errors."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonPartialOutputOverridesThrow" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP73 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 1, "flags")
	a, ak := semanticquery.NativeFlag(ctx, arg, 512)
	b, bk := semanticquery.NativeFlag(ctx, arg, 4194304)
	if ak && bk && a && b {
		ctx.ReportNode(arg, message)
	}
}
