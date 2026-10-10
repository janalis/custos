// Package failedclosurebindinginvoked implements the native FailedClosureBindingInvoked inspection.
package failedclosurebindinginvoked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check closure binding before invoking it."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FailedClosureBindingInvoked" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	bound := semanticquery.NativeLocalValue(ctx, call.Name)
	var closure, object syntax.Expr
	switch b := bound.(type) {
	case *syntax.MethodCall:
		if !semanticquery.NativeMethod(ctx, b, "Closure", "bindTo") {
			return
		}
		closure = b.Var
		object = semanticquery.CallArgument(b.Args, 0, "newThis")
	case *syntax.StaticCall:
		cl, ok := b.Class.(*syntax.Name)
		name, ok2 := b.Name.(*syntax.Identifier)
		if !ok || !ok2 || !strings.EqualFold(ctx.Names().Class(cl.Value, cl.Span().Start), "Closure") || !strings.EqualFold(name.Value, "bind") {
			return
		}
		closure = semanticquery.CallArgument(b.Args, 0, "closure")
		object = semanticquery.CallArgument(b.Args, 1, "newThis")
	default:
		return
	}
	static := false
	switch c := semanticquery.NativeValue(ctx, closure).(type) {
	case *syntax.Closure:
		static = c.Static
	case *syntax.ArrowFunction:
		static = c.Static
	}
	if !static || object == nil || len(ctx.Types().Native().TypeOf(object).Classes()) == 0 || semanticquery.NativeLocalSentinelGuard(ctx, call.Name, "null") {
		return
	}
	ctx.ReportNode(call, message)
}
