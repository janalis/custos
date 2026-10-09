package notoptimalregularexpressions

import (
	"testing"
)

// FuzzNorePattern runs the pattern helpers of NotOptimalRegularExpressions
// on arbitrary literal contents: no panic, and D15 agrees with the
// reference implementation.
func FuzzNorePattern(f *testing.F) {
	for _, s := range []string{
		"/[a-z][a-z]+/i", "#(\\d+)*#", "{[0-9]{2}[0-9]{2}}u", "/(?:a|\\w+)+$/D", "~>.*?<~", "/[/", "\xff/x/\xfe",
		"/[ab]{1,2}[ab]*[ab]/", "(a)", "/[]]/",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		body, _, ok := noreSplitDelimiters(raw)
		if !ok {
			body = raw
		}
		r1, c1, ok1 := noreRepeatedClass(body)
		r2, c2, ok2 := noreRepeatedClassRef(body)
		if r1 != r2 || c1 != c2 || ok1 != ok2 {
			t.Fatalf("%q: got (%q,%q,%v), want (%q,%q,%v)", body, r1, c1, ok1, r2, c2, ok2)
		}
		if len(body) <= noreNestedMaxLen {
			noreNestedQuantifiers(body)
		}
		_ = noreNormalize(body)
		_ = noreUnescapeText(body)
	})
}
