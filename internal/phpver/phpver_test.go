package phpver

import "testing"

func TestParse(t *testing.T) {
	ok := map[string]Version{"": Default, " 8.3 ": PHP83, "5.3": PHP53, "8.5.1": PHP85, "7.4": PHP74} //nolint:gocritic // " 8.3 ": surrounding spaces are trimmed
	for s, want := range ok {
		if v, err := Parse(s); err != nil || v != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", s, v, err, want)
		}
	}
	for _, s := range []string{"8", "x.1", "8.x", "4.4", "9.0", "8.-1", "8.100", "5.2", "8.6"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) accepted", s)
		}
	}
	if PHP83.String() != "8.3" || !PHP83.AtLeast(PHP80) || !PHP80.Below(PHP83) {
		t.Fatal("helpers")
	}
	if MustParse("8.1") != PHP81 {
		t.Fatal("MustParse")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustParse must panic on invalid input")
		}
	}()
	MustParse("nope")
}
