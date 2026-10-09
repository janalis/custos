package astquery

// IsComparisonNumericString reports whether s is a numeric string per PHP's grammar:
// optional surrounding whitespace, an optional sign, then an integer,
// a decimal (`1.`, `.5`, `1.5`) or either with an exponent (`1e3`).
func IsComparisonNumericString(s string) bool {
	isWS := func(c byte) bool {
		return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
	}
	for s != "" && isWS(s[0]) {
		s = s[1:]
	}
	for s != "" && isWS(s[len(s)-1]) {
		s = s[:len(s)-1]
	}
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := func() int {
		j := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		return i - j
	}
	n := digits()
	if i < len(s) && s[i] == '.' {
		i++
		n += digits()
	}
	if n == 0 {
		return false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if digits() == 0 {
			return false
		}
	}
	return i == len(s)
}
