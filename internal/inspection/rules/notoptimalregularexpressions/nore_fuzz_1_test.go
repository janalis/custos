package notoptimalregularexpressions

import (
	"strings"
)

// noreRepeatedClassRef is the original (quadratic) D15 scan, kept as the
// reference for the linear implementation.
func noreRepeatedClassRef(s string) (run, class string, ok bool) {
	classAt := func(i int) (int, bool) {
		if i >= len(s) || s[i] != '[' {
			return 0, false
		}
		j := strings.IndexByte(s[i+1:], ']')
		if j <= 0 {
			return 0, false
		}
		return i + 1 + j + 1, true
	}
	quantAt := func(i int) int {
		if i >= len(s) {
			return i
		}
		switch s[i] {
		case '*', '+', '?':
			return i + 1
		case '{':
			if j := strings.IndexByte(s[i+1:], '}'); j > 0 {
				return i + 1 + j + 1
			}
		}
		return i
	}
	rep := func(i int) (int, string, bool) {
		e1, ok := classAt(i)
		if !ok {
			return 0, "", false
		}
		c := s[i:e1]
		j := quantAt(e1)
		if !strings.HasPrefix(s[j:], c) {
			return 0, "", false
		}
		return quantAt(j + len(c)), c, true
	}
	for i := 0; i < len(s); i++ {
		end, cl, ok := rep(i)
		if !ok {
			continue
		}
		for {
			e2, c2, ok := rep(end)
			if !ok {
				break
			}
			end, cl = e2, c2
		}
		return s[i:end], cl, true
	}
	return "", "", false
}
