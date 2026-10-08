package util

// IsNumericString reports whether s is a numeric string for PHP 8
// (is_numeric): optional leading and trailing whitespace, an optional sign,
// decimal digits with an optional fraction and exponent. Two numeric
// strings compare as numbers under ==, so `"0 " == "00"` holds.
func IsNumericString(s string) bool {
	ws := func(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' }
	i, n := 0, len(s)
	for i < n && ws(s[i]) {
		i++
	}
	for n > i && ws(s[n-1]) {
		n--
	}
	if i < n && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := func() int {
		start := i
		for i < n && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		return i - start
	}
	d := digits()
	if i < n && s[i] == '.' {
		i++
		d += digits()
	}
	if d == 0 {
		return false
	}
	if i < n && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < n && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if digits() == 0 {
			return false
		}
	}
	return i == n
}
