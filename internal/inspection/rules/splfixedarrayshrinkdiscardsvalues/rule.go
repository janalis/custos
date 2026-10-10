// Package splfixedarrayshrinkdiscardsvalues implements the native SplFixedArrayShrinkDiscardsValues inspection.
package splfixedarrayshrinkdiscardsvalues

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve populated entries before shrinking."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SplFixedArrayShrinkDiscardsValues" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SplFixedArray", "setSize") {
		return
	}
	size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 0, "size"))
	old, ok := semanticquery.NativeFixedArraySize(ctx, c.Var, c)
	if !known || !ok || size < 0 || size >= old {
		return
	}
	for key := range semanticquery.NativeFixedArraySlots(ctx, c.Var, c) {
		if key >= size {
			ctx.ReportNode(c, message)
			return
		}
	}
}
