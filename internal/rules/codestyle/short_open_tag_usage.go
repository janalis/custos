package codestyle

import (
	"custos/internal/analysis"
	"custos/internal/syntax"
)

// shortOpenTagUsage reports `<?` open tags (dependent on short_open_tag).
type shortOpenTagUsage struct{}

func init() { register(shortOpenTagUsage{}) }

const shortOpenTagUsageMsg = "Replace the short open tag '<?' with '<?php'."

func (shortOpenTagUsage) ID() string { return "ShortOpenTagUsage" }

func (shortOpenTagUsage) Kinds() []syntax.NodeKind { return nil }

func (shortOpenTagUsage) Check(*analysis.Context, syntax.Node) {}

func (shortOpenTagUsage) CheckFile(ctx *analysis.Context) {
	src := ctx.Src
	for _, t := range ctx.File.Tokens {
		if t.Kind != syntax.TOpenTag || t.End-t.Start != 2 { // D1: exactly `<?`
			continue
		}
		if int(t.End) < len(src) { // D2 / E2
			switch src[t.End] {
			case ' ', '\t', '\n', '\r':
			default:
				continue
			}
		}
		span := syntax.Span{Start: t.Start, End: t.End}
		ctx.Report(span, shortOpenTagUsageMsg, analysis.Fix{
			Title: "Use '<?php'",
			Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: "<?php"}} },
		})
	}
}
