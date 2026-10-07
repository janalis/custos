package util

import "custos/internal/syntax"

// QuotedStringRaw returns the raw (undecoded) contents between the quotes of
// a single- or double-quoted string literal without interpolation, and the
// quote character. ok is false for any other node (heredoc/nowdoc, numbers,
// interpolated strings). A binary-string prefix (`b'…'`) is accepted.
func QuotedStringRaw(n syntax.Node) (content string, quote byte, ok bool) {
	lit, isLit := n.(*syntax.Literal)
	if !isLit || lit.LitKind != syntax.LitString {
		return "", 0, false
	}
	raw := lit.Raw
	if len(raw) > 0 && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if len(raw) < 2 || (raw[0] != '\'' && raw[0] != '"') || raw[len(raw)-1] != raw[0] {
		return "", 0, false
	}
	return raw[1 : len(raw)-1], raw[0], true
}
