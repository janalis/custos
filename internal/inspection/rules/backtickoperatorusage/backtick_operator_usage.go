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
	if span.Start >= span.End || span.Len() <= 2 || span.End > uint32(len(ctx.Src)) || ctx.Src[span.Start] != '`' || ctx.Src[span.End-1] != '`' { // D2/E1
		return
	}
	// Recovery may mistake an escaped final backtick for the closing
	// delimiter. Every part must lie inside the actual command content;
	// otherwise neither reporting nor slicing a replacement is justified.
	pos := span.Start + 1
	for _, part := range s.Parts {
		ps := part.Span()
		if ps.Start < pos || ps.End < ps.Start || ps.End > span.End-1 {
			return
		}
		pos = ps.End
	}
	src := ctx.Src
	fn := semanticquery.QualifiedBuiltin(ctx, "shell_exec", span.Start) // a namespaced shell_exec() would capture a bare call
	ctx.Report(span, backtickOperatorUsageMsg, diagnostic.Fix{
		Title: "Replace with shell_exec()",
		Edits: func() []diagnostic.TextEdit {
			repl := fn + "(\"" + backtickContent(src, s) + "\")"
			if span.Start > 0 && isIdentByte(src[span.Start-1]) {
				repl = " " + repl
			}
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// backtickContent converts the text between the backticks to double-quoted
// string content. Only literal parts are rewritten: interpolated expressions
// (and the braces around them) are PHP code and are copied untouched.
func backtickContent(src []byte, s *syntax.InterpolatedString) string {
	var b strings.Builder
	span := s.Span()
	pos := span.Start + 1
	for _, p := range s.Parts {
		if _, ok := p.(*syntax.StringPart); !ok {
			continue
		}
		ps := p.Span()
		b.Write(src[pos:ps.Start])
		writeLiteral(&b, src[ps.Start:ps.End])
		pos = ps.End
	}
	b.Write(src[pos : span.End-1])
	return b.String()
}

// writeLiteral re-escapes literal backtick text for a double-quoted string:
// an escaped backtick loses its backslash and a bare double quote gains one.
// Backticks do not treat backslash-quote as an escape, so that pair keeps its
// backslash: both characters are escaped.
func writeLiteral(b *strings.Builder, raw []byte) {
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '\\' && i+1 < len(raw):
			switch raw[i+1] {
			case '`':
				b.WriteByte('`')
			case '"':
				b.WriteString(`\\\"`)
			default:
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
}

func isIdentByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80
}
