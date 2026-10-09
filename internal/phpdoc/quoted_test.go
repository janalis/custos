package phpdoc

import (
	"strings"
	"testing"
)

func TestSplitTypeQuoted(t *testing.T) {
	for _, c := range []struct{ text, typ, rest string }{
		{`array{'first name': int, 'a}b': string} $row`, `array{'first name':int,'a}b':string}`, `$row`},
		{`array{"a>b,(c)[d]": int} $row`, `array{"a>b,(c)[d]":int}`, `$row`},
		{`'two  words' | "a\" | b" description`, `'two  words'|"a\" | b"`, `description`},
		{`array{'a\' }, b': int, 'slash\\': string} $row`, `array{'a\' }, b':int,'slash\\':string}`, `$row`},
		{"array{'tab\tkey':\tint}\n$row", "array{'tab\tkey':int}", "$row"},
		{`(T is 'two  words' ? array{'a  b': int} : null) desc`, `(T is 'two  words' ? array{'a  b': int} : null)`, `desc`},
		{`'unfinished words $row`, `'unfinished words $row`, ``},
		{`'unfinished\`, `'unfinished\`, ``},
		{`array{'unfinished } $row`, `array{'unfinished } $row`, ``},
		{`'`, `'`, ``},
		{`int`, `int`, ``},
	} {
		t.Run(c.text, func(t *testing.T) {
			typ, rest := SplitType(c.text)
			if typ != c.typ || rest != c.rest {
				t.Fatalf("got %q, %q; want %q, %q", typ, rest, c.typ, c.rest)
			}
		})
	}
}

func TestSquashQuotedWhitespace(t *testing.T) {
	for _, c := range []struct{ text, want string }{
		{" \tint\r\n|\vstring\f ", "int|string"},
		{"  (T is int ? ' a  b ' : string)  ", "(T is int ? ' a  b ' : string)"},
		{`' '`, `' '`},
		{`'a\' b'`, `'a\' b'`},
		{`'unfinished  words\`, `'unfinished  words\`},
	} {
		if got := squashType(c.text); got != c.want {
			t.Errorf("%q: got %q; want %q", c.text, got, c.want)
		}
	}
}

func TestQuotedAliasWhitespace(t *testing.T) {
	d := Parse("/** @phpstan-type Row array{'first  name': int, 'tab\tkey': string} */")
	if got := d.TypeAliases()["Row"]; got != "array{'first  name': int, 'tab\tkey': string}" {
		t.Fatalf("alias changes quoted whitespace: %q", got)
	}
}

func TestQuotedUnicodeWhitespace(t *testing.T) {
	for _, c := range []struct{ text, want string }{
		{"\u00a0array{'first\u00a0\u2003name':\u2003int}\u0085", "array{'first\u00a0\u2003name': int}"},
		{"(T is int ?\u2003' a\u00a0b ' :\u00a0string)", "(T is int ? ' a\u00a0b ' : string)"},
		{"array{é:\u2003int}", "array{é: int}"},
		{"array{\xff: int}", "array{\xff: int}"},
		{"int", "int"},
	} {
		if got := normalizeType(c.text, true); got != c.want {
			t.Errorf("%q: got %q; want %q", c.text, got, c.want)
		}
	}
	d := Parse("/** @phpstan-type Row array{'first\u00a0name':\u2003int} */")
	if got, want := d.TypeAliases()["Row"], "array{'first\u00a0name': int}"; got != want {
		t.Errorf("Unicode alias: got %q; want %q", got, want)
	}
	if got := squashType("(T is int ?\u2003' a\u00a0b ' :\u00a0string)"); got != "(T is int ? ' a\u00a0b ' : string)" {
		t.Errorf("Unicode conditional: %q", got)
	}
	if got := normalizeType("int\u00a0", false); got != "int\u00a0" {
		t.Errorf("nonconditional Unicode normalization changed: %q", got)
	}
}

func BenchmarkSplitTypeQuoted(b *testing.B) {
	for name, text := range map[string]string{
		"shape":        `array{'first name': int, 'a\' }, b': array{"x>y": string}} $row`,
		"conditional":  `(T is 'two  words' ? array{'a  b': int} : null) description`,
		"unterminated": "array{'" + strings.Repeat(`a\' b, } `, 1000),
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				SplitType(text)
			}
		})
	}
}

func BenchmarkNormalizeTypeWhitespace(b *testing.B) {
	for name, text := range map[string]string{
		"ascii":         `array{'first name': int, other: string}`,
		"unicode":       "array{'first\u00a0name':\u2003int, é:\u00a0string}",
		"no_whitespace": `array{id:int}`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				normalizeType(text, true)
			}
		})
	}
}
