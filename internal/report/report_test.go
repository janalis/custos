package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormats(t *testing.T) {
	items := []Item{{Path: "a.php", Line: 2, Column: 3, EndLine: 2, EndColumn: 5, Rule: "UnnecessarySemicolon", Severity: "info", Message: "m", Fixable: true}}
	for _, f := range Formats {
		var b bytes.Buffer
		if err := Write(&b, f, items, 1); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if !strings.Contains(b.String(), "a.php") {
			t.Errorf("%s output lacks path:\n%s", f, b.String())
		}
		if f == "json" || f == "sarif" {
			var v any
			if err := json.Unmarshal(b.Bytes(), &v); err != nil {
				t.Errorf("%s: invalid JSON: %v", f, err)
			}
		}
	}
	if err := Write(&bytes.Buffer{}, "nope", nil, 0); err == nil {
		t.Error("unknown format must fail")
	}
}
