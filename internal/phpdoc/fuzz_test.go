package phpdoc

import (
	"strings"
	"testing"
	"time"

	"custos/internal/testbudget"
)

// exercise runs every accessor on a parsed doc comment.
func exercise(c string) {
	d := Parse(c)
	d.Params()
	d.ReturnType()
	d.VarType("")
	d.VarType("x")
	d.TemplateParams()
	d.Templates()
	d.TypeAliases()
	for _, t := range d.Tags {
		SplitType(t.Text)
		_ = VarName(t.Text)
		if t.Name == "method" {
			ParseMethod(t.Text)
		}
	}
}

// FuzzParse checks that doc comment parsing never panics.
func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"/** @param int $x */", "/**\n * @return array<int, string> desc\n * @var $x Foo\n */",
		"/** @template T of array{a: int} */", "/** @phpstan-type A = int|string */",
		"/** @phpstan-import-type A from B as C */", "/** @param int |  null $x */", "/**", "*/",
		"/** @method static self make(array<int, string> &$out = ['k' => 'a, b'], string &...$rest) */",
		"/** @method callable(string): int callback(callable(int, string): bool $filter) */",
		"/** @method void broken(array<int] $row) */",
		`/** @param array{'first  name': int, 'a\' },b': string} $row */`,
		`/** @return (T is 'a ? b' ? array{'x}y': int} : null) */`,
		`/** @phpstan-type Row array{'tab key': int} */`,
		`/** @param 'unfinished\ */`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) { exercise(s) })
}

// TestParsePathological guards the linear-time behaviour on hostile doc
// comments (long tags, long blank runs in types, many tags).
func TestParsePathological(t *testing.T) {
	for name, c := range map[string]string{
		"long tag":      "/** @param int $x\n" + strings.Repeat(" * more text here\n", 200000) + " */",
		"blank run":     "/** @var a|" + strings.Repeat(" ", 1<<20) + "b */",
		"many tags":     "/**" + strings.Repeat(" * @param int $x\n", 100000) + "*/",
		"nesting":       "/** @return " + strings.Repeat("array<", 100000) + " */",
		"invalid utf-8": "/** @param \xff\xfe $\xff " + strings.Repeat("\xc3", 100000) + " */",
	} {
		start := time.Now()
		exercise(c)
		if d := time.Since(start); d > testbudget.Of(3*time.Second) {
			t.Errorf("%s: %v", name, d)
		}
	}
}
