// Package stringpademptypadding implements the native StringPadEmptyPadding inspection.
package stringpademptypadding

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonempty padding string."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StringPadEmptyPadding" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "str_pad") {
		return
	}
	pad, pk := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 2, "pad_string"))
	length, lk := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "length"))
	input, ik := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "string"))
	if pk && pad == "" && lk && ik && length > int64(len(input)) {
		ctx.ReportNode(c, message)
	}
}
