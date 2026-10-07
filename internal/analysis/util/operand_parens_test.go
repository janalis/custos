package util

import "testing"

func TestNeedsParensAsEqualityOperand(t *testing.T) {
	cases := map[string]bool{
		"$a;":              false,
		"$a . $b;":         false,
		"$a + $b;":         false,
		"$a < $b;":         false,
		"$a instanceof B;": false,
		"f($a);":           false,
		"!$a;":             false,
		"$a ?: $b;":        true,
		"$a ? $b : $c;":    true,
		"$a ?? $b;":        true,
		"$a === $b;":       true,
		"$a && $b;":        true,
		"$a | $b;":         true,
		"$a = $b;":         true,
		"$a .= $b;":        true,
	}
	for src, want := range cases {
		f := parse(t, "<?php "+src)
		if got := NeedsParensAsEqualityOperand(firstExpr(t, f)); got != want {
			t.Errorf("%s: got %v, want %v", src, got, want)
		}
	}
}
