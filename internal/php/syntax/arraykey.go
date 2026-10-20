package syntax

import (
	"strconv"
	"strings"
)

// LiteralArrayKey returns the canonical array key of a literal key expression
// (integer or plain string literal, optionally negated): "0", "-1", "name".
func LiteralArrayKey(x Expr) (string, bool) {
	switch n := UnwrapParens(x).(type) {
	case *Literal:
		switch n.LitKind {
		case LitInt:
			v, err := strconv.ParseInt(strings.ReplaceAll(n.Raw, "_", ""), 0, 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatInt(v, 10), true
		case LitString:
			s, ok := PlainStringLiteral(n.Raw)
			if !ok {
				return "", false
			}
			return s, true // "12" is the integer key 12, same canonical text
		}
	case *Unary:
		if n.Op.Kind == TMinus {
			if l, ok := UnwrapParens(n.Expr).(*Literal); ok && l.LitKind == LitInt {
				if k, ok := LiteralArrayKey(l); ok && k != "0" {
					return "-" + k, true
				}
			}
		}
	}
	return "", false
}

// PlainStringLiteral decodes a quoted string literal without escapes that need
// interpretation beyond `\\` and `\'` in single quotes. Narrower than
// StringLiteralValue on purpose: a double-quoted key with escapes or
// `$` is not given a shape key.
func PlainStringLiteral(raw string) (string, bool) {
	if raw != "" && (raw[0] == 'b' || raw[0] == 'B') {
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
		if strings.IndexByte(body, '\\') < 0 && strings.IndexByte(body, '$') < 0 {
			return body, true
		}
	}
	return "", false
}
