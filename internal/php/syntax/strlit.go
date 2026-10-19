package syntax

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	phpversion "custos/internal/php/version"
)

// StringLiteralValue decodes a quoted PHP string literal without interpolation.
// Heredoc, nowdoc and non-string source return false.
func StringLiteralValue(raw string) (val string, ok bool) {
	return StringLiteralValueForVersion(raw, phpversion.Max)
}

// StringLiteralValueForVersion decodes quoted strings using the selected PHP
// escape rules. Version zero selects the latest supported version.
func StringLiteralValueForVersion(raw string, version phpversion.Version) (val string, ok bool) {
	if version == 0 {
		version = phpversion.Max
	}
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
		return decodeDoubleQuotedForVersion(body, version)
	}
	return "", false
}

// decodeDoubleQuotedForVersion decodes the escapes of a double-quoted body.
// From PHP 7, a `\u{` escape without hexadecimal digits and a closing brace,
// or naming no Unicode code point, does not compile: ok is then false.
func decodeDoubleQuotedForVersion(s string, version phpversion.Version) (string, bool) {
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
			if version < phpversion.PHP54 {
				b.WriteString(`\e`)
			} else {
				b.WriteByte(0x1b)
			}
		case 'f':
			b.WriteByte('\f')
		case '\\', '$', '"':
			b.WriteByte(n)
		case 'x', 'X': // PHP accepts both introducers
			j := i + 2
			for j < len(s) && j < i+4 && isHex(s[j]) {
				j++
			}
			if j == i+2 {
				b.WriteByte('\\')
				b.WriteByte(n)
			} else {
				v, _ := strconv.ParseUint(s[i+2:j], 16, 8)
				b.WriteByte(byte(v))
			}
			i = j - 1
			continue
		case 'u':
			if version >= phpversion.PHP70 && i+2 < len(s) && s[i+2] == '{' {
				// Only hex digits may precede the '}' (scanning to the next
				// '}' per escape was quadratic on many unterminated `\u{`).
				end := 0
				for i+3+end < len(s) && isHex(s[i+3+end]) {
					end++
				}
				v, err := strconv.ParseUint(s[i+3:i+3+end], 16, 32)
				if end == 0 || i+3+end >= len(s) || s[i+3+end] != '}' || err != nil || v > unicode.MaxRune {
					return "", false
				}
				b.WriteRune(rune(v))
				i = i + 3 + end
				continue
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
	return b.String(), true
}

// LiteralStringValue decodes a parsed, noninterpolated string, including a
// complete heredoc or nowdoc. Its owning file supplies the PHP version.
func LiteralStringValue(literal *Literal) (string, bool) {
	if literal == nil || literal.LitKind != LitString {
		return "", false
	}
	version := SourceVersion(literal)
	if value, ok := StringLiteralValueForVersion(literal.Raw, version); ok {
		return value, true
	}
	return staticHeredocValueForVersion(literal.Raw, version)
}

// SourceVersion returns the parser version of a node's owning file. Detached
// nodes and files without an explicit version use the latest supported PHP.
func SourceVersion(node Node) phpversion.Version {
	for ; node != nil; node = node.Parent() {
		if value := reflect.ValueOf(node); value.Kind() == reflect.Pointer && value.IsNil() {
			break
		}
		if file, ok := node.(*File); ok {
			if file.Version != 0 {
				return file.Version
			}
			break
		}
	}
	return phpversion.Max
}
