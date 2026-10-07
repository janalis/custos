package util

import (
	"strconv"
	"strings"
)

// StringLiteralValue decodes the raw source of a single- or double-quoted
// string literal without interpolation (quotes included) according to PHP's
// escape rules for its quote style. ok is false for other literals
// (heredoc/nowdoc, numbers).
func StringLiteralValue(raw string) (val string, ok bool) {
	if len(raw) >= 1 && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if len(raw) < 2 || raw[len(raw)-1] != raw[0] {
		return "", false
	}
	body := raw[1 : len(raw)-1]
	switch raw[0] {
	case '\'':
		if strings.IndexByte(body, '\\') < 0 {
			return body, true
		}
		var b strings.Builder
		for i := 0; i < len(body); i++ {
			if body[i] == '\\' && i+1 < len(body) && (body[i+1] == '\\' || body[i+1] == '\'') {
				i++
			}
			b.WriteByte(body[i])
		}
		return b.String(), true
	case '"':
		if strings.IndexByte(body, '\\') < 0 {
			return body, true
		}
		return decodeDoubleQuoted(body), true
	}
	return "", false
}

func decodeDoubleQuoted(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		n := s[i+1]
		switch n {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'v':
			b.WriteByte('\v')
		case 'e':
			b.WriteByte(0x1b)
		case 'f':
			b.WriteByte('\f')
		case '\\', '$', '"':
			b.WriteByte(n)
		case 'x':
			j := i + 2
			for j < len(s) && j < i+4 && isHex(s[j]) {
				j++
			}
			if j == i+2 {
				b.WriteString(`\x`)
			} else {
				v, _ := strconv.ParseUint(s[i+2:j], 16, 8)
				b.WriteByte(byte(v))
			}
			i = j - 1
			continue
		case 'u':
			if i+2 < len(s) && s[i+2] == '{' {
				// Only hex digits may precede the '}' (scanning to the next
				// '}' per escape was quadratic on many unterminated `\u{`).
				end := 0
				for i+3+end < len(s) && isHex(s[i+3+end]) {
					end++
				}
				if end > 0 && i+3+end < len(s) && s[i+3+end] == '}' {
					if v, err := strconv.ParseUint(s[i+3:i+3+end], 16, 32); err == nil {
						b.WriteRune(rune(v))
						i = i + 3 + end
						continue
					}
				}
			}
			b.WriteString(`\u`)
		default:
			if n >= '0' && n <= '7' {
				j := i + 1
				for j < len(s) && j < i+4 && s[j] >= '0' && s[j] <= '7' {
					j++
				}
				v, _ := strconv.ParseUint(s[i+1:j], 8, 16)
				b.WriteByte(byte(v))
				i = j - 1
				continue
			}
			b.WriteByte('\\')
			b.WriteByte(n)
		}
		i++
	}
	return b.String()
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
