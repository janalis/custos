package syntax

import (
	"strconv"
	"strings"
	"unicode/utf8"

	phpversion "custos/internal/php/version"
)

// DecodedString pairs a constant PHP string with the source range of each
// decoded byte. Bytes produced by one escape share its entire source range.
type DecodedString struct {
	Value   string
	Sources []Span
}

// DecodeLiteralString decodes a complete noninterpolated literal and preserves
// its byte mapping, including indentation removal in flexible heredocs.
func DecodeLiteralString(literal *Literal) (DecodedString, bool) {
	value, ok := LiteralStringValue(literal)
	if !ok {
		return DecodedString{}, false
	}
	version := SourceVersion(literal)
	raw, base := literal.Raw, literal.Span().Start
	if raw[0] == 'b' || raw[0] == 'B' {
		raw, base = raw[1:], base+1
	}
	var body string
	var positions []uint32
	quote := raw[0]
	if quote == '\'' || quote == '"' {
		body = raw[1 : len(raw)-1]
		positions = make([]uint32, len(body))
		for i := range positions {
			positions[i] = base + 1 + uint32(i)
		}
	} else {
		l := lexer{src: []byte(raw), ver: version, states: []lexState{stScripting}, halt: -1}
		l.next() // LiteralStringValue already validated this heredoc.
		start, info := l.pos, l.heredocs[0]
		end := start
		for l.next() != TEndHeredoc {
			end = l.pos
		}
		text := strings.TrimSuffix(strings.TrimSuffix(raw[start:end], "\n"), "\r")
		indent := l.pos - len(info.label) - end
		var b strings.Builder
		skip := indent
		for i := 0; i < len(text); i++ {
			c := text[i]
			if skip > 0 && (c == ' ' || c == '\t') {
				skip--
				continue
			}
			b.WriteByte(c)
			positions = append(positions, base+uint32(start+i))
			if c == '\n' || c == '\r' {
				skip = indent
			} else {
				skip = 0
			}
		}
		body = b.String()
		quote = '"'
		if info.nowdoc {
			quote = 0
		}
	}
	out := DecodedString{Value: value, Sources: make([]Span, 0, len(value))}
	for i := 0; i < len(body); {
		end := stringEscapeEnd(body, i, quote, version)
		size := end - i
		if quote != 0 && body[i] == '\\' && size > 1 {
			n := body[i+1]
			switch {
			case quote == '\'':
				size = 1
			case n == 'u' && size > 2:
				code, _ := strconv.ParseUint(body[i+3:end-1], 16, 32)
				size = utf8.RuneLen(rune(code))
				if size < 0 { // The shared decoder emits a replacement rune.
					size = 3
				}
			case n == 'x' || n == 'X':
				if size > 2 {
					size = 1
				}
			case n >= '0' && n <= '7', strings.ContainsRune("ntrvf\\$\"", rune(n)), n == 'e' && version >= phpversion.PHP54:
				size = 1
			}
		}
		span := Span{Start: positions[i], End: positions[end-1] + 1}
		for range size {
			out.Sources = append(out.Sources, span)
		}
		i = end
	}
	return out, true
}

func stringEscapeEnd(body string, i int, quote byte, version phpversion.Version) int {
	if quote == 0 || body[i] != '\\' || i+1 == len(body) {
		return i + 1
	}
	n := body[i+1]
	if quote == '\'' {
		if n == '\\' || n == '\'' {
			return i + 2
		}
		return i + 1
	}
	end := i + 2
	switch {
	case n == 'x' || n == 'X':
		for end < len(body) && end < i+4 && isHex(body[end]) {
			end++
		}
	case n >= '0' && n <= '7':
		for end < len(body) && end < i+4 && body[end] >= '0' && body[end] <= '7' {
			end++
		}
	case n == 'u' && version >= phpversion.PHP70 && end < len(body) && body[end] == '{':
		for body[end] != '}' {
			end++ // The whole literal was validated before mapping.
		}
		end++
	}
	return end
}
