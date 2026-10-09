package datetimeconstantsusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// dateTimeConstantsUsage reports the non-compliant ISO8601 date format
// constants and offers the ATOM equivalents.
type dateTimeConstantsUsage struct{}

func (dateTimeConstantsUsage) ID() string { return "DateTimeConstantsUsage" }
func (dateTimeConstantsUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KConstFetch, syntax.KClassConstFetch}
}

func (dateTimeConstantsUsage) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.ConstFetch: // D1
		if semanticquery.GlobalConstName(ctx, x) != "DATE_ISO8601" {
			return
		}
		span := x.Name.Span()
		repl := `\DATE_ATOM`
		if !strings.HasPrefix(x.Name.Value, `\`) {
			repl = semanticquery.QualifiedGlobalConst(ctx, "DATE_ATOM", span.Start) // F1: a namespaced DATE_ATOM would capture a bare name
		}
		ctx.ReportNode(x, "DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.", diagnostic.Fix{
			Title: "Use DATE_ATOM",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
		})
	case *syntax.ClassConstFetch: // D2
		id, ok := x.Name.(*syntax.Identifier)
		if !ok || id.Value != "ISO8601" {
			return
		}
		found := false // D3
		for _, cls := range semanticquery.ReferencedClasses(ctx, x.Class) {
			if k := ctx.Index().FindConst(cls, "ISO8601", ctx.PHP); k != nil {
				switch strings.ToLower(strings.TrimPrefix(k.Class, `\`)) {
				case "datetime", "datetimeinterface", "datetimeimmutable":
					found = true
				}
				break
			}
		}
		if !found {
			return
		}
		span := id.Span()
		ctx.ReportNode(x, "ISO8601 is not ISO-8601 compliant; use the ATOM constant.", diagnostic.Fix{
			Title: "Use ATOM",
			Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: "ATOM"}} },
		})
	}
}

// Semantic marks the rule as needing the project index.
func (dateTimeConstantsUsage) Semantic() {}
