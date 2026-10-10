// Package mysqliinvalidbindingtype implements the native MysqliInvalidBindingType inspection.
package mysqliinvalidbindingtype

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use only supported mysqli binding type letters."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliInvalidBindingType" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli_stmt", "bind_param") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "types")
	types, known := semanticquery.NativeString(ctx, arg)
	if !known {
		return
	}
	for _, ch := range types {
		if ch != 'i' && ch != 'd' && ch != 's' && ch != 'b' {
			ctx.ReportNode(arg, message)
			return
		}
	}
}
