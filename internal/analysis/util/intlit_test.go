package util

import "testing"

func TestParseIntLiteral(t *testing.T) {
	ok := map[string]int64{"0": 0, "12": 12, "010": 8, "0o17": 15, "0O7": 7, "0x1F": 31, "0X0a": 10, "0b101": 5, "1_000": 1000, "0x1_f": 31, "0_7": 7}
	for in, want := range ok {
		if got, k := ParseIntLiteral(in); !k || got != want {
			t.Errorf("%q: got %d,%v want %d", in, got, k, want)
		}
	}
	for _, in := range []string{"", "08", "1.0", "1e3", "_1", "1_", "1__0", "0x_ff", "0x", "0b2", "-1", "99999999999999999999"} {
		if _, k := ParseIntLiteral(in); k {
			t.Errorf("%q: unexpectedly parsed", in)
		}
	}
}
