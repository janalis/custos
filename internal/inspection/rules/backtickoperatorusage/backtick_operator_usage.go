package backtickoperatorusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// backtickOperatorUsage reports shell-command backtick expressions and
// rewrites them as explicit shell_exec() calls.
type backtickOperatorUsage struct{}

const backtickOperatorUsageMsg = "Run the command through shell_exec() instead of backticks."

func (backtickOperatorUsage) ID() string { return "BacktickOperatorUsage" }
func (backtickOperatorUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInterpolatedString}
}

func (backtickOperatorUsage) Check(ctx *analysis.Context, n syntax.Node) {
	s, ok := n.(*syntax.InterpolatedString)
	if !ok || !s.Backtick {
		return
	}
	span := s.Span()
	if span.Len() <= 2 || ctx.Src[span.Start] != '`' || ctx.Src[span.End-1] != '`' { // D2/E1
		return
	}
	src := ctx.Src
	fn := semanticquery.QualifiedBuiltin(ctx, "shell_exec", span.Start) // a namespaced shell_exec() would capture a bare call
	ctx.Report(span, backtickOperatorUsageMsg, diagnostic.Fix{
		Title: "Replace with shell_exec()",
		Edits: func() []diagnostic.TextEdit {
			repl := fn + "(\"" + backtickContent(string(src[span.Start+1:span.End-1])) + "\")"
			if span.Start > 0 && isIdentByte(src[span.Start-1]) {
				repl = " " + repl
			}
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// backtickContent converts raw backtick content to double-quoted string
// content: `\“ becomes a plain backtick, an unescaped `"` becomes `\"`.
func backtickContent(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '\\' && i+1 < len(raw):
			if raw[i+1] == '`' {
				b.WriteByte('`')
			} else {
				b.WriteByte(c)
				b.WriteByte(raw[i+1])
			}
			i++
		case c == '"':
			b.WriteString(`\"`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func isIdentByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80
}
