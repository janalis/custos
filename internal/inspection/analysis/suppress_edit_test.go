package analysis

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestSuppressEdit(t *testing.T) {
	cases := []struct {
		name, src, target, want string // want: edited source, "" = no edit
	}{
		{
			name:   "nested statement gets its own line",
			src:    "<?php\nfunction f($a) {\n    $x = 1;\n    echo $a;\n}\n",
			target: "echo $a",
			want:   "<?php\nfunction f($a) {\n    $x = 1;\n    // @custos-ignore R\n    echo $a;\n}\n",
		},
		{
			name:   "statement sharing a line falls back to the enclosing one",
			src:    "<?php\n$y = 0;\nif ($a) { $x = 1; echo $a; }\n",
			target: "echo $a",
			want:   "<?php\n$y = 0;\n// @custos-ignore R\nif ($a) { $x = 1; echo $a; }\n",
		},
		{
			name:   "walks up through expressions (closure) to the statement",
			src:    "<?php\n$y = 0;\n$f = function () { echo $a; };\n",
			target: "echo $a",
			want:   "<?php\n$y = 0;\n// @custos-ignore R\n$f = function () { echo $a; };\n",
		},
		{
			name:   "tab indentation is kept",
			src:    "<?php\n$y = 0;\nwhile ($a) {\n\t$a--;\n}\n",
			target: "$a--",
			want:   "<?php\n$y = 0;\nwhile ($a) {\n\t// @custos-ignore R\n\t$a--;\n}\n",
		},
		{
			name:   "first statement of the file is never annotated (file-wide)",
			src:    "<?php\necho $a;\n",
			target: "echo $a",
		},
		{
			name:   "code on the opening-tag line has no line to annotate",
			src:    "<?php $y = 0; echo $a;\n",
			target: "echo $a",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := syntax.Parse("x.php", []byte(c.src), syntax.Options{Version: phpversion.PHP84})
			at := strings.Index(c.src, c.target)
			span := syntax.Span{Start: uint32(at), End: uint32(at + len(c.target))}
			e, ok := SuppressEdit(f, "R", span)
			if c.want == "" {
				if ok {
					t.Fatalf("unexpected edit %+v", e)
				}
				return
			}
			if !ok {
				t.Fatal("no edit")
			}
			got := c.src[:e.Span.Start] + e.NewText + c.src[e.Span.End:]
			if got != c.want {
				t.Errorf("got\n%s\nwant\n%s", got, c.want)
			}
		})
	}
}
