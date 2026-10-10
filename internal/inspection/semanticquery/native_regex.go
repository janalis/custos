package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// NativeRegexCaptures counts ordinary and named groups in a conservative
// literal PCRE subset. Unsupported control constructs remain unknown.
func NativeRegexCaptures(pattern string) (int, bool) {
	if len(pattern) < 2 {
		return 0, false
	}
	delimiter := pattern[0]
	if delimiter >= 'a' && delimiter <= 'z' || delimiter >= 'A' && delimiter <= 'Z' || delimiter >= '0' && delimiter <= '9' || delimiter == '\\' || delimiter <= ' ' {
		return 0, false
	}
	if strings.ContainsRune("({[<", rune(delimiter)) {
		return 0, false
	}
	end := strings.LastIndexByte(pattern, delimiter)
	if end == 0 {
		return 0, false
	}
	for _, modifier := range pattern[end+1:] {
		if !strings.ContainsRune("imsADSUXJur", modifier) {
			return 0, false
		}
	}
	count, depth := 0, 0
	class := false
	for i := 1; i < end; i++ {
		c := pattern[i]
		if c == '\\' {
			if i+1 >= end {
				return 0, false
			}
			// Quoted spans and control escapes consume bytes that this
			// small counter cannot safely classify as group syntax.
			if i+1 < end && strings.ContainsRune("QEc", rune(pattern[i+1])) {
				return 0, false
			}
			i++
			continue
		}
		if c == delimiter {
			return 0, false
		}
		if class && c == '[' {
			return 0, false
		}
		if c == '[' && !class {
			if i+1 < end && (pattern[i+1] == ']' || pattern[i+1] == '^' && i+2 < end && pattern[i+2] == ']') {
				return 0, false
			}
			class = true
			continue
		}
		if c == ']' && class {
			class = false
			continue
		}
		if class {
			continue
		}
		if c == '(' {
			if i+1 < end && pattern[i+1] == '*' {
				return 0, false
			}
			depth++
			if i+1 < end && pattern[i+1] == '?' {
				if i+2 >= end {
					return 0, false
				}
				switch pattern[i+2] {
				case ':', '=', '!':
				case '<':
					if i+3 >= end {
						return 0, false
					}
					if pattern[i+3] != '=' && pattern[i+3] != '!' {
						if strings.IndexByte(pattern[i+3:end], '>') < 1 {
							return 0, false
						}
						count++
					}
				case '\'':
					count++
				case 'P':
					if i+3 >= end || pattern[i+3] != '<' {
						return 0, false
					}
					count++
				default:
					return 0, false
				}
			} else {
				count++
			}
		}
		if c == ')' {
			depth--
			if depth < 0 {
				return 0, false
			}
		}
	}
	return count, depth == 0 && !class
}

// NativeFlagContains resolves a pure named flag or bitwise-or tree.
func NativeFlagContains(ctx *analysis.Context, e syntax.Expr, name string) (bool, bool) {
	e = syntax.UnwrapParens(e)
	if c, ok := e.(*syntax.ConstFetch); ok {
		return GlobalConstName(ctx, c) == name, GlobalConstName(ctx, c) != ""
	}
	if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		left, lk := NativeFlagContains(ctx, b.Left, name)
		right, rk := NativeFlagContains(ctx, b.Right, name)
		return left || right, lk && rk
	}
	if value, ok := NativeInt(ctx, e); ok {
		var bit int64
		switch name {
		case "PREG_SPLIT_DELIM_CAPTURE":
			bit = 2
		case "PREG_OFFSET_CAPTURE":
			bit = 256
		default:
			return false, false
		}
		return value&bit != 0, true
	}
	return false, false
}
