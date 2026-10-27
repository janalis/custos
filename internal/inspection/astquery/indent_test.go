package astquery

import (
	"testing"

	"custos/internal/php/syntax"
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

func TestStepIndent(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// nested under a loop indented less: the difference
		{"<?php\nfunction g() {\n\tif (1) {}\n        for (;;) {\n            echo 1;\n        }\n}", "    "},
		// tabs
		{"<?php\nfunction g() {\n\tfor (;;) {\n\t\techo 1;\n\t}\n}", "\t"},
		// an enclosing node on the same indentation is passed over
		{"<?php\nif (1) {\n  if (2) {} else\n  if (3) {\n      echo 1;\n  }\n}", "    "},
		// enclosing indentation that is not a prefix: the file's guess
		{"<?php\nfunction g() {\n\tfor (;;) {\n    echo 1;\n\t}\n}", "    "},
		// top level: the file's guess
		{"<?php\n  echo 1;", "  "},
	} {
		f := syntax.Parse("t.php", []byte(tc.src), syntax.Options{})
		var echo syntax.Node
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if _, ok := n.(*syntax.Echo); ok && echo == nil {
				echo = n
			}
			return true
		})
		indent := LineIndent(f.Src, echo.Span().Start)
		if got := StepIndent(f, echo, indent); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.src, got, tc.want)
		}
	}
	if got := LineStart([]byte("ab\r\ncd"), 5); got != 4 {
		t.Errorf("LineStart: got %d", got)
	}
}
