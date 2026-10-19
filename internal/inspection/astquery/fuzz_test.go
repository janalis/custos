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
	// An unterminated or malformed `\u{` escape does not compile from PHP 7:
	// it is rejected, as fast as the others are decoded.
	for name, tc := range map[string]struct {
		raw     string
		decoded bool
	}{
		"unterminated \\u{": {`"` + strings.Repeat(`\u{`, 300000) + `"`, false},
		"long hex":          {`"\u{` + strings.Repeat("0", 1<<20) + `41}"`, true},
		"octal/hex runs":    {`"` + strings.Repeat(`\x\777\u`, 200000) + `"`, true},
		"single escapes":    {`'` + strings.Repeat(`\'\\`, 300000) + `'`, true},
	} {
		start := time.Now()
		if _, ok := StringLiteralValue(tc.raw); ok != tc.decoded {
			t.Errorf("%s: decoded=%v", name, ok)
		}
		if d := time.Since(start); d > testbudget.Of(2*time.Second) {
			t.Errorf("%s: %v", name, d)
		}
	}
	if v, _ := StringLiteralValue(`"\u{` + strings.Repeat("0", 1000) + `41}"`); v != "A" {
		t.Errorf("leading zeros: %q", v)
	}
	if _, ok := StringLiteralValue(`"\u{4 1}\u{}"`); ok {
		t.Error("malformed escape decoded")
	}
	if _, ok := StringLiteralValue(`"\u{110000}"`); ok {
		t.Error("code point beyond Unicode decoded")
	}
}
