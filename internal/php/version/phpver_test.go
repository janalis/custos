package version

import "testing"

func TestParse(t *testing.T) {
	ok := map[string]Version{"": Default, " 8.3 ": PHP83, "5.3": PHP53, "8.5.1": PHP85, "7.4": PHP74} //nolint:gocritic // " 8.3 ": surrounding spaces are trimmed
	for s, want := range ok {
		if v, err := Parse(s); err != nil || v != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", s, v, err, want)
		}
	}
	for _, s := range []string{"8", "x.1", "8.x", "4.4", "9.0", "8.-1", "8.100", "5.2", "8.6", "5.7", "5.99", "6.0", "6.99", "7.5", "7.99"} {
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

func TestNext(t *testing.T) {
	for v, want := range map[Version]Version{699: PHP70, PHP56: PHP70, PHP74: PHP80, 799: PHP80, PHP84: PHP85} {
		if got, ok := Next(v); got != want || !ok {
			t.Errorf("Next(%v) = %v, %v", v, got, ok)
		}
	}
	if _, ok := Next(Max); ok || Releases[0] != Min || Releases[len(Releases)-1] != Max {
		t.Error("releases bounds")
	}
}
