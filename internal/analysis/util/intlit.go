package util

import (
	"strconv"
	"strings"
)

// ParseIntLiteral parses the source text of a PHP integer literal the way
// PHP does: decimal, octal with a leading `0` or `0o`/`0O`, hexadecimal
// `0x`, binary `0b`, and `_` digit separators between digits. ok is false
// for anything else (floats, malformed separators, invalid digits) or when
// the value overflows int64 (PHP would make it a float).
func ParseIntLiteral(s string) (int64, bool) {
	base, digits := 10, s
	if len(s) > 1 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			base, digits = 16, s[2:]
		case 'b', 'B':
			base, digits = 2, s[2:]
		case 'o', 'O':
			base, digits = 8, s[2:]
		default: // legacy octal: the leading 0 is itself a digit (`0_7`)
			base = 8
		}
	}
	if digits == "" || digits[0] == '_' || digits[len(digits)-1] == '_' || strings.Contains(digits, "__") {
		return 0, false
	}
	digits = strings.ReplaceAll(digits, "_", "")
	if digits[0] == '+' || digits[0] == '-' {
		return 0, false
	}
	v, err := strconv.ParseInt(digits, base, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
