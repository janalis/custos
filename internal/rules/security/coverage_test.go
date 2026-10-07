package security

import (
	"strings"
	"testing"
)

// TestParseJSONRejects covers every syntax-error exit of the composer.json
// parser (malformed manifests are skipped by SecurityAdvisories).
func TestParseJSONRejects(t *testing.T) {
	for _, s := range []string{
		"",                       // empty input
		strings.Repeat("[", 600), // nesting limit
		`{1: 2}`,                 // key not a string
		`{"a" 1}`,                // missing colon
		`{"a": }`,                // bad member value
		`{"a": 1`,                // unterminated object
		`{"a": 1 "b": 2}`,        // missing comma
		`[1, }`,                  // bad array item
		`[1`,                     // unterminated array
		`[1 2]`,                  // missing comma
		`"abc`,                   // unterminated string
		`x`,                      // not a value
		`{} x`,                   // trailing data
	} {
		if _, ok := parseJSON([]byte(s)); ok {
			t.Errorf("%q: accepted", s)
		}
	}
	if v, ok := parseJSON([]byte(`[true, false, null, -1.5e3, "a\/b", "\x"]`)); !ok || len(v.items) != 6 || v.items[4].str != "a/b" || v.items[5].str != `\x` {
		t.Errorf("valid array: %+v %v", v, ok)
	}
}

func TestSecurityAdvisoriesFilePatterns(t *testing.T) {
	if got := (securityAdvisories{}).FilePatterns(); len(got) != 1 || got[0] != "composer.json" {
		t.Errorf("FilePatterns() = %v", got)
	}
}
