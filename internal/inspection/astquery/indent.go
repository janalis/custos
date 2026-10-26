package astquery

import (
	"strings"

	"custos/internal/php/syntax"
)

// LineIndent returns the leading spaces and tabs of the line containing
// byte offset off.
func LineIndent(src []byte, off uint32) string {
	if int(off) > len(src) {
		off = uint32(len(src))
	}
	start := int(off)
	for start > 0 && src[start-1] != '\n' && src[start-1] != '\r' {
		start--
	}
	end := start
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}
	return string(src[start:end])
}

// IndentBefore returns the horizontal whitespace between the start of the
// line and off, or "" when other code precedes off on its line (unlike
// LineIndent, which returns the line's indentation regardless).
func IndentBefore(src []byte, off uint32) string {
	i := int(off)
	for i > 0 && (src[i-1] == ' ' || src[i-1] == '\t') {
		i--
	}
	if i > 0 && src[i-1] != '\n' && src[i-1] != '\r' {
		return ""
	}
	return string(src[i:off])
}

// IsSpace reports whether c is a space, tab, newline or carriage return.
func IsSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// NameByte reports whether c can be part of a PHP name or keyword, so that
// text placed right next to it would merge with it.
func NameByte(c byte) bool {
	return c == '_' || c >= 0x80 || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// Reindent returns the source of span with each line after its first moved
// from the indentation of the span's first line to indent (lines not
// starting with that indentation are kept). It keeps the text when a line
// break lies inside a string, heredoc or inline HTML, whose content would
// change.
func Reindent(f *syntax.File, span syntax.Span, indent string) string {
	text := string(f.Src[span.Start:span.End])
	if !strings.Contains(text, "\n") {
		return text
	}
	for i := TokenIndex(f, span.Start); i < len(f.Tokens); i++ {
		t := f.Tokens[i]
		if t.Start >= span.End {
			break
		}
		if t.Kind != syntax.TWhitespace && t.Kind != syntax.TComment && t.Kind != syntax.TDocComment &&
			strings.Contains(string(f.Src[t.Start:t.End]), "\n") {
			return text
		}
	}
	base := LineIndent(f.Src, span.Start)
	lines := strings.Split(text, "\n")
	for i := 1; i < len(lines); i++ {
		if rest, ok := strings.CutPrefix(lines[i], base); ok {
			lines[i] = indent + rest
		}
	}
	return strings.Join(lines, "\n")
}
