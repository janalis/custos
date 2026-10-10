// Package monthadditiondateoverflow implements the MonthAdditionDateOverflow inspection.
package monthadditiondateoverflow

import (
	"strings"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MonthAdditionDateOverflow" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }

const message = "Choose an explicit month-end overflow policy."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "modify") || semanticquery.NativeDateClass(ctx, call.Var) == "" {
		return
	}
	change, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "modifier"))
	if !known || (change != "+1 month" && change != "-1 month") {
		return
	}
	value, ok := semanticquery.NativeValue(ctx, call.Var).(*syntax.New)
	if !ok {
		return
	}
	if semanticquery.NativeDateClass(ctx, call.Var) == "DateTime" {
		if _, fresh := syntax.UnwrapParens(call.Var).(*syntax.New); !fresh {
			return
		}
	}
	date, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(value.Args, 0, "datetime"))
	if !known {
		return
	}
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return
	}
	delta := 1
	if change == "-1 month" {
		delta = -1
	}
	month := time.Date(day.Year(), day.Month()+time.Month(delta), 1, 0, 0, 0, 0, time.UTC)
	last := month.AddDate(0, 1, -1).Day()
	if day.Day() > last {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
