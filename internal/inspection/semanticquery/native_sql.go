package semanticquery

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// SQLTokens tokenizes a deliberately bounded SQL subset. Strings and
// comments never become keywords or placeholders; uncertain syntax is refused.
func SQLTokens(sql string) ([]string, bool) {
	if len(sql) > 65536 {
		return nil, false
	}
	var out []string
	for i := 0; i < len(sql); {
		ch := sql[i]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '$' {
			return nil, false
		}
		if ch == '#' || (ch == '-' && i+1 < len(sql) && sql[i+1] == '-') {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if ch == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 || strings.Contains(sql[i+2:i+2+end], "/*") {
				return nil, false
			}
			i += end + 4
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			q := ch
			i++
			var text strings.Builder
			closed := false
			for i < len(sql) {
				if sql[i] == '\\' {
					if i+1 >= len(sql) {
						return nil, false
					}
					text.WriteByte(sql[i+1])
					i += 2
					continue
				}
				if sql[i] == q {
					if i+1 < len(sql) && sql[i+1] == q {
						text.WriteByte(q)
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				text.WriteByte(sql[i])
				i++
			}
			if !closed {
				return nil, false
			}
			prefix := "'"
			if q == '`' {
				prefix = "`"
			}
			out = append(out, prefix+text.String())
			continue
		}
		if ch == ':' && i+1 < len(sql) && sql[i+1] == ':' {
			out = append(out, "::")
			i += 2
			continue
		}
		if ch == ':' || ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			start := i
			i++
			for i < len(sql) {
				c := sql[i]
				if c != '_' && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') {
					break
				}
				i++
			}
			word := sql[start:i]
			if ch != ':' {
				word = strings.ToUpper(word)
			}
			out = append(out, word)
			continue
		}
		out = append(out, string(ch))
		i++
	}
	return out, true
}

// PDODriver reports whether e holds a PDO created with a DSN for driver.
func PDODriver(ctx *analysis.Context, e syntax.Expr, driver string) bool {
	construct := NativeConstruction(ctx, e, "PDO")
	if construct == nil {
		return false
	}
	dsn, known := NativeString(ctx, CallArgument(construct.Args, 0, "dsn"))
	return known && strings.HasPrefix(strings.ToLower(dsn), driver+":")
}

// SQLSkeletonParts splits a constant SQL skeleton around dynamic PHP operands.
// It accepts only literal concatenation and ordinary interpolated strings.
func SQLSkeletonParts(ctx *analysis.Context, e syntax.Expr) ([]string, []syntax.Expr, bool) {
	var text strings.Builder
	var parts []string
	var values []syntax.Expr
	var visit func(syntax.Expr) bool
	visit = func(e syntax.Expr) bool {
		e = syntax.UnwrapParens(e)
		if s, ok := NativeString(ctx, e); ok {
			text.WriteString(s)
			return true
		}
		if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			return visit(b.Left) && visit(b.Right)
		}
		if s, ok := e.(*syntax.InterpolatedString); ok {
			if s.Backtick {
				return false
			}
			for _, p := range s.Parts {
				if literal, ok := p.(*syntax.StringPart); ok {
					text.WriteString(literal.Raw)
				} else {
					parts = append(parts, text.String())
					text.Reset()
					values = append(values, p)
				}
			}
			return true
		}
		parts = append(parts, text.String())
		text.Reset()
		values = append(values, e)
		return true
	}
	if !visit(e) {
		return nil, nil, false
	}
	parts = append(parts, text.String())
	return parts, values, len(values) > 0
}
