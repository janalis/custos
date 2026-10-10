// Package mysqliplaceholderbindingmismatch implements the native MysqliPlaceholderBindingMismatch inspection.
package mysqliplaceholderbindingmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Bind one value for each SQL placeholder."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliPlaceholderBindingMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli_stmt", "bind_param") || c.Args == nil {
		return
	}
	producer, ok := semanticquery.NativeValue(ctx, c.Var).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "mysqli", "prepare") {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(producer.Args, 0, "query"))
	if !known {
		return
	}
	count, known := semanticquery.NativeSQLPositionalCount(sql)
	if !known {
		return
	}
	args := 0
	for _, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil {
			return
		}
		args++
	}
	if count != args-1 {
		ctx.ReportNode(c, message)
	}
}
