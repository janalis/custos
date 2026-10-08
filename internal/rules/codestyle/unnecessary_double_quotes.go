package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// unNecessaryDoubleQuotes reports double-quoted strings without
// interpolation or escapes that could be single-quoted.
type unNecessaryDoubleQuotes struct{}

func init() { register(unNecessaryDoubleQuotes{}) }

const unNecessaryDoubleQuotesMsg = "No interpolation or escapes here; use single quotes."

var dqUnescaper = strings.NewReplacer(`\$`, `$`, `\"`, `"`)

func (unNecessaryDoubleQuotes) ID() string { return "UnNecessaryDoubleQuotes" }

func (unNecessaryDoubleQuotes) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KLiteral}
}

func (unNecessaryDoubleQuotes) Check(ctx *analysis.Context, n syntax.Node) {
	lit := n.(*syntax.Literal)
	if lit.LitKind != syntax.LitString || n.Span().Len() == 0 {
		return
	}
	raw := lit.Raw
	prefix := ""
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') { // binary prefix kept
		prefix, raw = raw[:1], raw[1:]
	}
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' { // D1
		return
	}
	content := dqUnescaper.Replace(raw[1 : len(raw)-1])
	if strings.ContainsAny(content, `'\`) { // D3
		return
	}
	span := n.Span()
	ctx.Report(span, unNecessaryDoubleQuotesMsg, analysis.Fix{
		Title: "Use single quotes",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: prefix + "'" + content + "'"}}
		},
	})
}
