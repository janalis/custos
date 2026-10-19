package types

import (
	"math"
	"testing"
)

// String lengths ride on the string atom like int values, refine from any
// length, and fall back to unconstrained when every length is possible.
func TestStringLengths(t *testing.T) {
	read := Union(StringLengths(IntRange{0, 4}), Of("false"))
	for _, tc := range []struct {
		name string
		got  Type
		want string
	}{
		{"seed", read, "false|string<len 0, 4>"},
		{"exact", read.LengthsWithin(4, 4), "false|string<len 4, 4>"},
		{"not empty", read.LengthsWithout(0), "false|string<len 1, 4>"},
		{"emptied", read.LengthsWithin(5, 9), "false"},
		{"plain refined", String.LengthsWithin(16, 16), "string<len 16, 16>"},
		{"plain not empty", String.LengthsWithout(0), "string<len 1, max>"},
		{"back to any", StringLengths(IntRange{0, 3}, IntRange{4, math.MaxInt64}), "string"},
		{"negative lengths", StringLengths(IntRange{-5, -1}), "?unknown"},
		{"union", Union(StringLengths(IntRange{1, 1}), StringLengths(IntRange{4, 4})), "string<len 1, 1>|string<len 4, 4>"},
		{"union with plain", Union(StringLengths(IntRange{1, 1}), String), "string"},
		{"without other", read.Without("false"), "string<len 0, 4>"},
		{"also", read.LengthsAlso(StringLengths(IntRange{3, 9})), "false|string<len 3, 4>"},
		{"also plain", read.LengthsAlso(String), "false|string<len 0, 4>"},
		{"also no string", read.LengthsAlso(Of("false")), "false"},
		{"no string", Of("false").LengthsWithin(1, 1).LengthsWithout(0).LengthsAlso(String), "false"},
		{"relative", Union(StringLengths(IntRange{2, 2}), Of("static")).BindRelative(`\A`, "", `\A`), `\A|string<len 2, 2>`},
	} {
		if got := tc.got.ShapeString(); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
	if !StringLengths(IntRange{4, 4}).LengthsOnly(4, 4) || read.LengthsOnly(0, 4) || String.LengthsOnly(0, 9) {
		t.Error("LengthsOnly")
	}
	if _, ok := String.StringLengthRanges(); ok {
		t.Error("a plain string is unconstrained")
	}
}
