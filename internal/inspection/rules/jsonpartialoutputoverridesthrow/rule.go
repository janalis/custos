// Package jsonpartialoutputoverridesthrow implements the native JsonPartialOutputOverridesThrow inspection.
package jsonpartialoutputoverridesthrow

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
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
	if ctx.PHP.Below(phpversion.PHP73) {
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
