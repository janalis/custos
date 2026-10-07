package util

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
