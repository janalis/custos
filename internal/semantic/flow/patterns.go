package flow

import (
	"regexp"
	"strings"
)

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
	body := pattern[1:end]
	if flags != "" {
		body = "(?" + flags + ")" + body
	}
	re, err := regexp.Compile(body)
	return err == nil && !re.MatchString("")
}
