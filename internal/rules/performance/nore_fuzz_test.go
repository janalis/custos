package performance

import (
	"strings"
	"testing"
	"time"
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

// TestNorePathological guards against super-linear pattern scans.
func TestNorePathological(t *testing.T) {
	for name, body := range map[string]string{
		"open brackets":   strings.Repeat("[", 1<<20),
		"open braces":     strings.Repeat("[a]{", 300000),
		"nested classes":  strings.Repeat("[", 300000) + "]" + strings.Repeat("[", 300000) + "]",
		"many groups":     strings.Repeat("a(bc)d", 2000) + strings.Repeat("x", 1<<20),
		"repeated":        strings.Repeat("[ab]", 300000),
		"delimiters only": "/" + strings.Repeat("/", 1<<20),
	} {
		start := time.Now()
		noreSplitDelimiters(body)
		noreRepeatedClass(body)
		if len(body) <= noreNestedMaxLen {
			noreNestedQuantifiers(body)
		}
		if d := time.Since(start); d > 3*time.Second {
			t.Errorf("%s: %v", name, d)
		}
	}
	// the largest patterns D18 still examines stay fast: deep nesting and
	// many distinct groups
	k := noreNestedMaxLen / 4
	nested := ("x" + strings.Repeat("a(", k) + "bc" + strings.Repeat(")d", k))[:noreNestedMaxLen]
	var b strings.Builder
	for i := 0; b.Len() < noreNestedMaxLen-16; i++ {
		b.WriteString("a(b" + string(rune('0'+i%10)) + string(rune('A'+i/10%26)) + string(rune('a'+i/260%26)) + ")c")
	}
	for _, p := range []string{nested, b.String()} {
		start := time.Now()
		noreNestedQuantifiers(p)
		if d := time.Since(start); d > 3*time.Second {
			t.Errorf("max-size D18 pattern: %v", d)
		}
	}
	nest := func(k int) string { return `x(\d+)+y` + strings.Repeat("(", k) + "ab" + strings.Repeat(")c", k) }
	if got := noreNestedQuantifiers(nest(10)); len(got) != 1 {
		t.Errorf("below the fold cap: got %v", got)
	}
	if got := noreNestedQuantifiers(nest(noreNestedMaxFolds + 1)); got != nil {
		t.Errorf("fold cap: got %v", got)
	}
}
