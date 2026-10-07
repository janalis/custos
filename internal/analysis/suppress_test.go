package analysis

import (
	"strings"
	"testing"
	"time"

	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/testbudget"
)

func analyseOK(t *testing.T, src string) []Finding {
	t.Helper()
	e, err := NewEngine([]Rule{okRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84}))
}

func TestSuppressions(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"none", "<?php\n$a = !$b;\n$c = !$d;", 2},
		{"file level", "<?php\n// @noinspection NestedNotOperators\n$a = !$b;\n$c = !$d;", 0},
		{"statement", "<?php\nf();\n/** @noinspection NestedNotOperatorsInspection */\n$a = !$b;\n$c = !$d;", 1},
		{"custos-ignore list", "<?php\nf();\n// @custos-ignore Foo, NestedNotOperators\n$a = !$b;\n$c = !$d;", 1},
		{"all", "<?php\nf();\n/* @noinspection ALL */\n$a = !$b;\n$c = !$d;", 1},
		{"other rule", "<?php\nf();\n// @noinspection Foo\n$a = !$b;", 1},
		{"enclosing function", "<?php\nf();\n/** @noinspection NestedNotOperators */\nfunction g() { if (!$x) { return !$y; } }\n$c = !$d;", 1},
		{"closure statement", "<?php\nf();\n$f = function () {\n  // @noinspection NestedNotOperators\n  return !$x;\n};\n$g = function () { return !$y; };", 1},
		{"not directly before", "<?php\nf();\n// @noinspection NestedNotOperators\ng();\n$a = !$b;", 1},
		{"param", "<?php\nf();\nfunction g(\n  /** @noinspection NestedNotOperators */ $p = !X,\n  $q = !Y) {}", 1},
		{"sibling statements", "<?php\nf();\nfunction g() {\n  // @noinspection NestedNotOperators\n  $a = !$b;\n  // @noinspection NestedNotOperators\n  $c = !$d;\n  $e = !$f;\n}", 1},
		{"nested", "<?php\nf();\n// @noinspection Foo\nfunction g() {\n  // @noinspection NestedNotOperators\n  $a = !$b;\n  $c = !$d;\n}", 1},
		{"second tag same line", "<?php\nf();\n// @noinspection Foo @noinspection NestedNotOperators\n$a = !$b;", 0},
	} {
		if got := analyseOK(t, tc.src); len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.name, len(got), tc.want, got)
		}
	}
}

// TestSuppressionsScale guards against the quadratic suppression lookup
// (every finding used to re-walk the whole tree from the top) and the
// quadratic tag parsing of a long single-line comment.
func TestSuppressionsScale(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nf();\n")
	const n = 40000
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			b.WriteString("// @noinspection NestedNotOperators\n")
		}
		b.WriteString("$a = !$b;\n")
	}
	// Below the engine's findings cap: drive the suppression index directly.
	f := syntax.Parse("x.php", []byte(b.String()), syntax.Options{Version: phpver.PHP84})
	start := time.Now()
	sup := newSuppressions(f)
	kept := 0
	syntax.InspectFile(f, func(nd syntax.Node) bool {
		if nd.Kind() == syntax.KUnary && !sup.suppressed(Finding{Rule: "NestedNotOperators", Span: nd.Span()}) {
			kept++
		}
		return true
	})
	if kept != n/2 {
		t.Fatalf("kept %d findings, want %d", kept, n/2)
	}
	if d := time.Since(start); d > testbudget.Of(5*time.Second) && !raceEnabled {
		t.Errorf("%d findings took %v", n, d)
	}

	line := "// " + strings.Repeat("@noinspection ", 200000)
	start = time.Now()
	ids := ParseSuppressionComment(line)
	if d := time.Since(start); d > testbudget.Of(2*time.Second) && !raceEnabled {
		t.Errorf("long comment took %v", d)
	}
	if len(ids) > 200000 {
		t.Errorf("long comment: %d ids", len(ids))
	}
}

// TestFindingsCap checks that a file with a huge number of findings is
// truncated with one note instead of collecting them all.
func TestFindingsCap(t *testing.T) {
	got := analyseOK(t, "<?php\n"+strings.Repeat("$a = !$b;\n", MaxFindingsPerFile+500))
	if len(got) != MaxFindingsPerFile+1 {
		t.Fatalf("got %d findings", len(got))
	}
	last := got[len(got)-1]
	if last.Rule != "internal" || !strings.Contains(last.Message, "more than") {
		t.Fatalf("last finding %+v", last)
	}
}

// FuzzParseSuppressionComment checks the suppression tag parser never
// panics and yields no more names than the comment has bytes.
func FuzzParseSuppressionComment(f *testing.F) {
	for _, s := range []string{"/** @noinspection A, B */", "// @custos-ignore X\n// @noinspection", "@noinspection@noinspection *", "/* @custos-ignore\r\nA */"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, c string) {
		if ids := ParseSuppressionComment(c); len(ids) > len(c) {
			t.Fatalf("%q: %d ids", c, len(ids))
		}
	})
}
