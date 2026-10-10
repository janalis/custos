// Package dateparsenormalizationunchecked implements the DateParseNormalizationUnchecked inspection.
package dateparsenormalizationunchecked

import (
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateParseNormalizationUnchecked" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }

const message = "Reject or explicitly accept normalized calendar dates."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := semanticquery.NativeDateStatic(ctx, n, "createFromFormat")
	if call == nil {
		return
	}
	format, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "format"))
	if !ok || (format != "!Y-m-d" && format != "Y-m-d") {
		return
	}
	date, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "datetime"))
	if !ok {
		return
	}
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
		return
	}
	for i, c := range date {
		if i != 4 && i != 7 && (c < '0' || c > '9') {
			return
		}
	}
	if _, err := time.Parse("2006-01-02", date); err == nil {
		return
	}
	// An explicit diagnostics check after the parse is an accepted validation contract.
	lastErrors := ctx.Memo("lastErrors", func() any {
		latest := map[syntax.Node]uint32{}
		syntax.InspectFile(ctx.File, func(node syntax.Node) bool {
			if semanticquery.NativeDateStatic(ctx, node, "getLastErrors") != nil {
				latest[syntax.EnclosingVariableScope(node)] = node.Span().Start
			}
			return true
		})
		return latest
	}).(map[syntax.Node]uint32)
	if lastErrors[syntax.EnclosingVariableScope(call)] <= call.Span().End {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
