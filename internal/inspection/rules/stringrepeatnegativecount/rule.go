// Package stringrepeatnegativecount implements the native StringRepeatNegativeCount inspection.
package stringrepeatnegativecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonnegative repetition count."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StringRepeatNegativeCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "str_repeat") {
		return
	}
	v, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "times"))
	if k && v < 0 {
		ctx.ReportNode(c, message)
	}
}
