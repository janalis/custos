package security

import (
	"strings"
	"testing"
	"time"

	"custos/internal/testbudget"
)

// FuzzParseJSON checks the composer.json parser of SecurityAdvisories: no
// panic, and the spans of an accepted document stay inside the input.
func FuzzParseJSON(f *testing.F) {
	for _, s := range []string{
		`{"name": "a/b", "require": {"php": ">=8.1", "x/y": "^1"}, "require-dev": {}}`,
		`[1, -2.5e3, true, false, null, "é\n"]`, `{"a": {"b": [{}]}}`, `"\`, `{"a":`, `{,}`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, ok := parseJSON([]byte(s))
		if !ok {
			return
		}
		var check func(v *jsonValue)
		check = func(v *jsonValue) {
			if v == nil || int(v.span.End) > len(s) || v.span.Start > v.span.End {
				t.Fatalf("%q: bad value %+v", s, v)
			}
			for _, m := range v.members {
				check(m.value)
			}
			for _, it := range v.items {
				check(it)
			}
		}
		check(v)
	})
}

func TestParseJSONPathological(t *testing.T) {
	for name, s := range map[string]string{
		"deep arrays":   strings.Repeat("[", 1000000),
		"deep objects":  strings.Repeat(`{"a":`, 300000),
		"long string":   `"` + strings.Repeat(`\"`, 1000000) + `"`,
		"many members":  "{" + strings.Repeat(`"k":1,`, 300000) + `"k":1}`,
		"open string":   `{"` + strings.Repeat("a", 1<<20),
		"trailing junk": `{}` + strings.Repeat(" ", 1<<20) + "x",
	} {
		start := time.Now()
		parseJSON([]byte(s))
		if d := time.Since(start); d > testbudget.Of(3*time.Second) {
			t.Errorf("%s: %v", name, d)
		}
	}
}
