// Package intervaldaycomponentastotal implements the IntervalDayComponentAsTotal inspection.
package intervaldaycomponentastotal

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntervalDayComponentAsTotal" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }

const message = "Use total days instead of the day component."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "format") {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 0, "format")
	format, ok := semanticquery.NativeString(ctx, arg)
	if !ok || format != "%d" {
		return
	}
	origin, ok := syntax.UnwrapParens(call.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	method, ok := origin.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(method.Value, "diff") || semanticquery.NativeDateClass(ctx, origin.Var) == "" {
		return
	}
	literal, ok := arg.(*syntax.Literal)
	if ok && literal.LitKind == syntax.LitString && len(literal.Raw) >= 2 && (literal.Raw[0] == '\'' || literal.Raw[0] == '"') {
		ctx.ReportNode(call, message, astquery.ReplaceFix(arg.Span(), literal.Raw[:1]+"%a"+literal.Raw[len(literal.Raw)-1:]))
		return
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
