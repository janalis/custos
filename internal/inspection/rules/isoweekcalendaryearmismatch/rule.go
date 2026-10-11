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

// calendarTokens are the format characters whose value belongs to the
// calendar year: month, day of month, day of year, month length, leap-year
// flag, and the composite ISO 8601, RFC 2822 and Unix timestamp forms.
const calendarTokens = "mnFMdjztLcrU"

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	arg := semanticquery.NativeDateFormat(ctx, n)
	format, ok := semanticquery.NativeString(ctx, arg)
	if !ok {
		return
	}
	tokens := semanticquery.NativeFormatTokens(format)
	if !tokens['Y'] || !tokens['W'] || tokens['o'] {
		return
	}
	// Y is the right year for any calendar-date part, so it only mislabels
	// the week when the format has no such part.
	for i := 0; i < len(calendarTokens); i++ {
		if tokens[calendarTokens[i]] {
			return
		}
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
