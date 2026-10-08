package util

import "testing"

func TestIsNumericString(t *testing.T) {
	for s, want := range map[string]bool{
		"0": true, "00": true, " 011 ": true, "-1.5": true, "+.5": true, "5.": true, "1e3": true, "1E-3": true,
		"": false, " ": false, "-": false, ".": false, "1e": false, "1e+": false, "0x1A": false, "abc": false, "12a": false, "get:": false,
	} {
		if got := IsNumericString(s); got != want {
			t.Errorf("IsNumericString(%q) = %v", s, got)
		}
	}
}
