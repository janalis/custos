package securityadvisories

import (
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

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
