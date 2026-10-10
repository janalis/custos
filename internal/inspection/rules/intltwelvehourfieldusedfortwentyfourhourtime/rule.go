// Package intltwelvehourfieldusedfortwentyfourhourtime implements the native IntlTwelveHourFieldUsedForTwentyFourHourTime inspection.
package intltwelvehourfieldusedfortwentyfourhourtime

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the 24-hour calendar field."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlTwelveHourFieldUsedForTwentyFourHourTime" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "IntlCalendar", "set") {
		return
	}
	field := semanticquery.CallArgument(c.Args, 0, "field")
	value, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "value"))
	if !ok || value < 12 || value > 23 || !semanticquery.ExpansionCConstant(ctx, field, "IntlCalendar", "FIELD_HOUR") {
		return
	}
	constant := syntax.UnwrapParens(field).(*syntax.ClassConstFetch)
	ctx.ReportNode(c, message, astquery.ReplaceFix(constant.Name.Span(), "FIELD_HOUR_OF_DAY"))
}
