// Package mysqlibindingtypecountmismatch implements the native MysqliBindingTypeCountMismatch inspection.
package mysqlibindingtypecountmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the binding type count to the supplied variables."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliBindingTypeCountMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli_stmt", "bind_param") || c.Args == nil {
		return
	}
	types, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "types"))
	if !known {
		return
	}
	count := 0
	for _, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil {
			return
		}
		count++
	}
	if len(types) != count-1 {
		ctx.ReportNode(c, message)
	}
}
