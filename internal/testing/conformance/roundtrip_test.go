package conformance

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// fixRoundTrips lists own fixtures held to a stricter fix check than
// TestOwnFixtures: their quick-fixes must produce `<name>.fixed.php` byte
// for byte, whitespace included, and leave nothing for the rule to report.
var fixRoundTrips = []struct {
	rule     string
	fixtures []string
}{
	{"ContentLengthUsesCharacterCount", []string{"basic"}},
	{"Formatting", []string{"basic"}},
	{"JsonIntegerPrecisionLoss", []string{"basic"}},
	{"NanCheckedWithEquality", []string{"basic"}},
	{"OpenSslVerifyTruthyResult", []string{"basic", "conditions"}},
	{"PasswordComparedWithFreshHash", []string{"basic"}},
	{"PdoExecuteArrayReplacesBindings", []string{"basic"}},
	{"RepeatedCookieHeaderReplacement", []string{"basic"}},
	{"SortComparatorReturnsBoolean", []string{"basic"}},
	{"SortComparatorReturnsFloat", []string{"basic"}},
	{"UsortDiscardsRequiredKeys", []string{"basic"}},
	{"ValidatedIntegerZeroRejected", []string{"basic", "options"}},
}

// TestOwnFixturesRoundTrip applies the fixes of each listed fixture in one
// pass and analyzes the result again. Its name matches `make fixtures`.
func TestOwnFixturesRoundTrip(t *testing.T) {
	for _, c := range fixRoundTrips {
		if !wanted([]string{c.rule}) {
			continue
		}
		dir := filepath.Join(ownRoot, strings.ToLower(c.rule), "testdata")
		for _, name := range c.fixtures {
			t.Run(c.rule+"/"+name, func(t *testing.T) {
				req, marked, fixed := ownCase(t, c.rule, dir, filepath.Join(dir, name+".php"), ".json")
				if fixed == nil {
					t.Fatalf("%s.fixed.php is missing", name)
				}
				clean, _, err := ParseMarkup(marked)
				if err != nil {
					t.Fatal(err)
				}
				req.Source = clean
				out, err := engine.Fix(req)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(out, fixed) {
					t.Fatalf("fixed output differs:\n--- want\n%s\n--- got\n%s", fixed, out)
				}
				req.Source = out
				again, err := engine.Analyze(req)
				if err != nil {
					t.Fatal(err)
				}
				if len(again) != 0 {
					t.Fatalf("the fixes left findings: %+v", again)
				}
			})
		}
	}
}
