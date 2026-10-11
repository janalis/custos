package substrusedasstrpos

import (
	"testing"
)

func TestIsNumericString(t *testing.T) {
	for s, want := range map[string]bool{
		"0": true, "00": true, " 011 ": true, "-1.5": true, "+.5": true, "5.": true, "1e3": true, "1E-3": true,
		"": false, " ": false, "-": false, ".": false, "1e": false, "1e+": false, "0x1A": false, "abc": false, "12a": false, "get:": false,
	} {
		if got := isNumericString(s); got != want {
			t.Errorf("isNumericString(%q) = %v", s, got)
		}
	}
}

func TestIsHexString(t *testing.T) {
	for s, want := range map[string]bool{
		"0x1A": true, "0Xff": true, "\t0x1": true, "-0x1": true,
		"": false, "0x": false, "0": false, "x1A": false, "10x1": false, "tel:": false,
	} {
		if got := isHexString(s); got != want {
			t.Errorf("isHexString(%q) = %v", s, got)
		}
	}
}
