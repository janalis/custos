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

func TestIndentUnit(t *testing.T) {
	for _, tc := range []struct{ src, indent, want string }{
		{"<?php\nif ($a) {\n    b();\n}", "", "    "},
		{"<?php\n/**\n * doc\n */\nif ($a) {\n  b();\n}", "  ", "  "},
		{"<?php\nf();\n\n", "", "    "},
		{"<?php\n  \t$x;\n                  $y;", "", "    "},
		{"<?php\n  $x;", "\t", "\t"},
	} {
		if got := IndentUnit([]byte(tc.src), tc.indent); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.src, got, tc.want)
		}
	}
}
