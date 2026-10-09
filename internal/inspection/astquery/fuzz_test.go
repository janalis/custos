package astquery

import (
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// FuzzStringLiteralValue checks escape decoding never panics and that a
// literal without escapes decodes to its body.
func FuzzStringLiteralValue(f *testing.F) {
	for _, s := range []string{`'a\'b'`, `"\x41\101\u{48}\e\$"`, `b"\u{"`, `"\u{110000}"`, `"\7777"`, `''`, `"`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		v, ok := StringLiteralValue(raw)
		if ok && !strings.Contains(raw, `\`) {
			body := strings.TrimLeft(raw, "bB")
			if v != body[1:len(body)-1] {
				t.Fatalf("%q decoded to %q", raw, v)
			}
		}
		ParseIntLiteral(raw)
	})
}

// TestStringLiteralValuePathological guards linear decoding on hostile
// literals (unterminated \u{ escapes used to rescan to the end each).
func TestStringLiteralValuePathological(t *testing.T) {
	for name, raw := range map[string]string{
		"unterminated \\u{": `"` + strings.Repeat(`\u{`, 300000) + `"`,
		"long hex":          `"\u{` + strings.Repeat("0", 1<<20) + `41}"`,
		"octal/hex runs":    `"` + strings.Repeat(`\x\777\u`, 200000) + `"`,
		"single escapes":    `'` + strings.Repeat(`\'\\`, 300000) + `'`,
	} {
		start := time.Now()
		if _, ok := StringLiteralValue(raw); !ok {
			t.Errorf("%s: not decoded", name)
		}
		if d := time.Since(start); d > testbudget.Of(2*time.Second) {
			t.Errorf("%s: %v", name, d)
		}
	}
	if v, _ := StringLiteralValue(`"\u{` + strings.Repeat("0", 1000) + `41}"`); v != "A" {
		t.Errorf("leading zeros: %q", v)
	}
	if v, _ := StringLiteralValue(`"\u{4 1}\u{}"`); v != `\u{4 1}\u{}` {
		t.Errorf("malformed: %q", v)
	}
}
