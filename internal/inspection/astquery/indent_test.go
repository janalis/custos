package astquery

import (
	"testing"
)

func TestLineIndent(t *testing.T) {
	src := []byte("a\n\t  if (x) {\n}\r\n  y")
	for off, want := range map[uint32]string{0: "", 5: "\t  ", 15: "", 19: "  ", 99: "  "} {
		if got := LineIndent(src, off); got != want {
			t.Errorf("%d: got %q want %q", off, got, want)
		}
	}
}
