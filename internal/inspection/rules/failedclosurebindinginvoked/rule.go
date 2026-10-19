// Package failedclosurebindinginvoked implements the native FailedClosureBindingInvoked inspection.
package failedclosurebindinginvoked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Check closure binding before invoking it."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FailedClosureBindingInvoked" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	for _, bound := range semanticquery.NativeLocalBuiltinResults(ctx, call.Name, "null", "Closure::bind", "Closure::bindTo") {
		if failedBinding(ctx, call, bound) {
			ctx.ReportNode(call, message)
			return
		}
	}
}

func failedBinding(ctx *analysis.Context, call *syntax.FuncCall, bound syntax.Expr) bool {
	var closure, object syntax.Expr
	switch b := bound.(type) {
	case *syntax.MethodCall:
		closure = b.Var
		object = semanticquery.CallArgument(b.Args, 0, "newThis")
	case *syntax.StaticCall:
		closure = semanticquery.CallArgument(b.Args, 0, "closure")
		object = semanticquery.CallArgument(b.Args, 1, "newThis")
	}
	static := false
	switch c := semanticquery.NativeValue(ctx, closure).(type) {
	case *syntax.Closure:
		static = c.Static
	case *syntax.ArrowFunction:
		static = c.Static
	}
	if !static || object == nil || !provenObject(ctx, object) || ctx.Flow().Excludes(call.Name, "null") || !semanticquery.NativeFailureFlowOnly(ctx, call.Name) && semanticquery.NativeNarrowedExcludes(ctx, call.Name, types.Of(`\Closure`, "null"), "null") {
		return false
	}
	return true
}

// provenObject requires every alternative of the bound value to be a class
// instance: binding null to a static closure succeeds.
func provenObject(ctx *analysis.Context, object syntax.Expr) bool {
	t := ctx.Types().Native().TypeOf(object)
	return len(t.Classes()) > 0 && len(t.Classes()) == len(t.Atoms())
}
