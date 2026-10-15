package syntax

import (
	"strings"

	phpversion "custos/internal/php/version"
)

// staticHeredocValue decodes only a complete noninterpolated heredoc/nowdoc.
// Reusing the lexer validates delimiters and indentation without growing AST
// nodes or broadening the public quoted-string decoder's contract.
func staticHeredocValue(raw string) (string, bool) {
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if !strings.HasPrefix(raw, "<<<") {
		return "", false
	}
	l := lexer{src: []byte(raw), ver: phpversion.Max, states: []lexState{stScripting}, halt: -1}
	if l.next() != TStartHeredoc {
		return "", false
	}
	start := l.pos
	info := l.heredocs[0]
	for l.pos < len(raw) {
		end := l.pos
		switch l.next() {
		case TEncapsedAndWhitespace:
			continue
		case TEndHeredoc:
			if l.pos != len(raw) || len(l.errs) != 0 {
				return "", false
			}
			body := raw[start:end]
			if strings.HasSuffix(body, "\r\n") {
				body = body[:len(body)-2]
			} else if strings.HasSuffix(body, "\n") || strings.HasSuffix(body, "\r") {
				body = body[:len(body)-1]
			}
			indent := raw[end : l.pos-len(info.label)]
			if indent != "" {
				var stripped strings.Builder
				stripped.Grow(len(body))
				skip := len(indent)
				for i := 0; i < len(body); i++ {
					c := body[i]
					if skip > 0 && c == indent[0] {
						skip--
						continue
					}
					stripped.WriteByte(c)
					if c == '\n' || c == '\r' {
						skip = len(indent)
					} else {
						skip = 0
					}
				}
				body = stripped.String()
			}
			if info.nowdoc {
				return body, true
			}
			if !strings.Contains(body, `\`) {
				return body, true
			}
			return decodeDoubleQuoted(body), true
		default:
			return "", false
		}
	}
	return "", false
}
