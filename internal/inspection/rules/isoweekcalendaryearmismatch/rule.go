// Package isoweekcalendaryearmismatch implements the IsoWeekCalendarYearMismatch inspection.
package isoweekcalendaryearmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IsoWeekCalendarYearMismatch" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

const message = "Use the ISO week year with the ISO week number."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	arg := semanticquery.NativeDateFormat(ctx, n)
	format, ok := semanticquery.NativeString(ctx, arg)
	if !ok {
		return
	}
	tokens := semanticquery.NativeFormatTokens(format)
	if !tokens['Y'] || !tokens['W'] || tokens['o'] || ((tokens['m'] || tokens['n'] || tokens['F'] || tokens['M']) && (tokens['d'] || tokens['j'])) {
		return
	}
	literal, ok := arg.(*syntax.Literal)
	if ok && literal.LitKind == syntax.LitString && len(literal.Raw) >= 2 && (literal.Raw[0] == '\'' || literal.Raw[0] == '"') && literal.Raw[1:len(literal.Raw)-1] == format {
		raw := []byte(literal.Raw)
		for i := 1; i < len(raw)-1; i++ {
			if raw[i] == '\\' {
				i++
				continue
			}
			if raw[i] == 'Y' {
				raw[i] = 'o'
			}
		}
		ctx.ReportNode(n, message, astquery.ReplaceFix(arg.Span(), string(raw)))
		return
	}
	ctx.ReportNode(n, message)
}

func (rule) Semantic() {}
