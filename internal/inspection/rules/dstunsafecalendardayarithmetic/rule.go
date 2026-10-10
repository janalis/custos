// Package dstunsafecalendardayarithmetic implements the DstUnsafeCalendarDayArithmetic inspection.
package dstunsafecalendardayarithmetic

import (
	"strings"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DstUnsafeCalendarDayArithmetic" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }

const message = "Use calendar arithmetic for the next local day."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	binary := n.(*syntax.Binary)
	if binary.Op.Kind != syntax.TPlus {
		return
	}
	stamp := binary.Left
	duration := binary.Right
	seconds, known := semanticquery.NativeInt(ctx, duration)
	if !known || seconds != 86400 {
		seconds, known = semanticquery.NativeInt(ctx, binary.Left)
		stamp = binary.Right
	}
	if !known || seconds != 86400 {
		return
	}
	call, ok := semanticquery.NativeValue(ctx, stamp).(*syntax.MethodCall)
	if !ok {
		return
	}
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "getTimestamp") || semanticquery.NativeDateClass(ctx, call.Var) == "" {
		return
	}
	date, ok := semanticquery.NativeValue(ctx, call.Var).(*syntax.New)
	if !ok {
		return
	}
	if semanticquery.NativeDateClass(ctx, call.Var) == "DateTime" {
		if _, fresh := syntax.UnwrapParens(call.Var).(*syntax.New); !fresh {
			return
		}
	}
	text, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(date.Args, 0, "datetime"))
	if !ok {
		return
	}
	if len(text) != 19 || text[10:] != " 00:00:00" {
		return
	}
	zone, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(date.Args, 1, "timezone")).(*syntax.New)
	if !ok || ctx.Types().ClassRef(zone.Class) != "DateTimeZone" {
		return
	}
	zoneName, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(zone.Args, 0, "timezone"))
	if !ok {
		return
	}
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		return
	}
	day, err := time.ParseInLocation("2006-01-02 15:04:05", text, location)
	if err != nil {
		return
	}
	_, winter := time.Date(day.Year(), 1, 1, 0, 0, 0, 0, location).Zone()
	_, summer := time.Date(day.Year(), 7, 1, 0, 0, 0, 0, location).Zone()
	if winter != summer {
		ctx.ReportNode(binary, message)
	}
}

func (rule) Semantic() {}
