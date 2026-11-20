package formatting

import (
	"slices"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// formatting reports each change the formatter makes to a file.
type formatting struct{}

func (formatting) ID() string                           { return "Formatting" }
func (formatting) Kinds() []syntax.NodeKind             { return nil }
func (formatting) Check(*analysis.Context, syntax.Node) {}

// Format marks the rule as reading the format settings of each file.
func (formatting) Format() {}

func (formatting) CheckFile(ctx *analysis.Context) {
	v, ok := ctx.FormatSettings()
	if !ok { // E1: left out of formatting
		return
	}
	settings, _ := v.(*format.Settings) // nil: the default settings
	res := format.Source(ctx.File.Path, ctx.Src, format.Options{Settings: settings, PHP: ctx.PHP})
	if res.Skipped != "" { // E2
		return
	}
	var all []diagnostic.TextEdit
	for _, c := range res.Changes {
		if !c.Report {
			all = append(all, diagnostic.TextEdit{Span: syntax.Span{Start: c.Start, End: c.End}, NewText: c.New})
		}
	}
	// F2: the whole layout at once. Callers may sort the edits they get.
	whole := diagnostic.Fix{Title: "Format the file", Edits: func() []diagnostic.TextEdit { return slices.Clone(all) }}
	for _, c := range res.Changes {
		span := syntax.Span{Start: c.Start, End: c.End}
		if c.Report { // D2: no fix
			ctx.Report(span, c.Message())
			continue
		}
		// D1, F1: this change alone.
		edit := []diagnostic.TextEdit{{Span: span, NewText: c.New}}
		ctx.Report(span, c.Message(), diagnostic.Fix{Title: c.Message(), Edits: func() []diagnostic.TextEdit { return slices.Clone(edit) }}, whole)
	}
}
