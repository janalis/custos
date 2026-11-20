package format

import (
	"fmt"
	"os"
	"testing"
)

// TestFuzzReplay shows how $CUSTOS_FORMAT_FILE formats twice with each
// preset (development aid for fuzz failures).
func TestFuzzReplay(t *testing.T) {
	path := os.Getenv("CUSTOS_FORMAT_FILE")
	if path == "" {
		t.Skip("CUSTOS_FORMAT_FILE not set")
	}
	src, _ := os.ReadFile(path)
	for _, p := range Presets() {
		r, _ := Resolve(p, true)
		one := Source(path, src, Options{Settings: &r.Settings})
		two := Source(path, one.Output, Options{Settings: &r.Settings})
		if string(one.Output) != string(two.Output) {
			fmt.Printf("== %s\n%q\n%q\n", p, one.Output, two.Output)
			for _, c := range two.Changes {
				fmt.Printf("  %d-%d %s %q\n", c.Start, c.End, c.Rule, c.New)
			}
		}
	}
}
