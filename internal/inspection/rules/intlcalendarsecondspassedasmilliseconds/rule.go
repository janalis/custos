// Package intlcalendarsecondspassedasmilliseconds implements the native IntlCalendarSecondsPassedAsMilliseconds inspection.
package intlcalendarsecondspassedasmilliseconds

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Convert seconds to calendar milliseconds."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntlCalendarSecondsPassedAsMilliseconds" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "IntlCalendar", "setTime") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "timestamp")
	origin, ok := semanticquery.NativeValue(ctx, arg).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, origin, "time") {
		return
	}
	if direct, ok := syntax.UnwrapParens(arg).(*syntax.FuncCall); ok {
		ctx.ReportNode(c, message, astquery.ReplaceFix(arg.Span(), "("+ctx.Text(direct)+" * 1000)"))
	} else {
		ctx.ReportNode(c, message)
	}
}
