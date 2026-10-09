package notoptimalregularexpressions

import (
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

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
		if d := time.Since(start); d > testbudget.Of(3*time.Second) {
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
		if d := time.Since(start); d > testbudget.Of(3*time.Second) {
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
