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
