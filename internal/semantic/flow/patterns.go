package flow

import (
	"regexp"
	"strings"

	"custos/internal/php/syntax"
)

// EmptylessPattern reports whether x is a literal PCRE pattern that can
// never match the empty text (see emptyless).
func EmptylessPattern(x syntax.Expr) bool {
	pattern, known := literalText(x)
	return known && emptyless(pattern)
}

// emptyless reports whether a literal PCRE pattern, delimiters and modifiers
// included, can never match the empty text. Only patterns Go's regular
// expressions read the same way qualify: lookarounds, backreferences,
// possessive or extended syntax and unknown escapes prove nothing.
func emptyless(pattern string) bool {
	if pattern == "" {
		return false
	}
	open := pattern[0]
	closing := open
	switch open {
	case '(':
		closing = ')'
	case '[':
		closing = ']'
	case '{':
		closing = '}'
	case '<':
		closing = '>'
	}
	if open == '\\' || open == ' ' || ('a' <= open|0x20 && open|0x20 <= 'z') || ('0' <= open && open <= '9') {
		return false
	}
	end := strings.LastIndexByte(pattern, closing)
	if end < 1 {
		return false
	}
	flags := ""
	for _, m := range pattern[end+1:] {
		switch m {
		case 'i', 'm', 's', 'U':
			flags += string(m)
		case 'u', 'D', 'S':
		default:
			return false
		}
	}
	body, ok := goEscapes(pattern[1:end])
	if !ok {
		return false
	}
	if flags != "" {
		body = "(?" + flags + ")" + body
	}
	re, err := regexp.Compile(body)
	return err == nil && !re.MatchString("")
}

// goEscapes rewrites the PCRE escapes Go's syntax lacks or reads otherwise
// into atoms that, like them, match exactly one character or line break
// sequence (`\R`, `\h`, `\N`, `\X` and the like), which is all the empty
// match test needs. In a character class, `\h` and `\v` become the spaces
// they stand for; an escape a class cannot hold proves nothing.
func goEscapes(body string) (string, bool) {
	var b strings.Builder
	class := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			i++
			switch e := body[i]; {
			case strings.IndexByte("RHVNXhv", e) < 0:
				b.WriteByte(c)
				b.WriteByte(e)
			case !class:
				b.WriteString("x")
			case e == 'h':
				b.WriteString(` \t`)
			case e == 'v':
				b.WriteString(`\n\v\f\r`)
			default:
				return "", false
			}
			continue
		case c == '[' && !class:
			class = true
			b.WriteByte(c)
			// A `]` (after an optional `^`) opening the class is literal.
			if i+1 < len(body) && body[i+1] == '^' {
				i++
				b.WriteByte('^')
			}
			if i+1 < len(body) && body[i+1] == ']' {
				i++
				b.WriteString(`\]`)
			}
			continue
		case c == ']' && class:
			class = false
		}
		b.WriteByte(c)
	}
	return b.String(), true
}
