package types

import (
	"slices"
	"strings"
	"testing"
)

func TestQuotedTypeScanners(t *testing.T) {
	for _, c := range []struct {
		text string
		sep  byte
		want []string
	}{
		{`'a\'|b'|int`, '|', []string{`'a\'|b'`, `int`}},
		{`"a\"|b"|int`, '|', []string{`"a\"|b"`, `int`}},
		{`'slash\\'|int`, '|', []string{`'slash\\'`, `int`}},
		{`'a,b',array{'x:y': int}`, ',', []string{`'a,b'`, `array{'x:y': int}`}},
		{`'a}b'|int`, '|', []string{`'a}b'`, `int`}},
		{`int|'unfinished|float`, '|', []string{`int`, `'unfinished|float`}},
		{`'unfinished\`, '|', []string{`'unfinished\`}},
	} {
		if got := splitTop(c.text, c.sep); !slices.Equal(got, c.want) {
			t.Errorf("splitTop(%q): %q; want %q", c.text, got, c.want)
		}
	}
	for _, text := range []string{
		`{'a}b': int}`, `<array{'x>y': int}>`, `('a)b')`,
		`{'a\'}b': int}`, `{"a\"}b": int}`, `{'slash\\': int}`,
	} {
		if got := matchingClose(text); got != len(text)-1 {
			t.Errorf("matchingClose(%q): %d; want %d", text, got, len(text)-1)
		}
	}
	for _, text := range []string{`{'unfinished}`, `{'unfinished\`, `{'`, ``, `{`} {
		if got := matchingClose(text); got != -1 {
			t.Errorf("matchingClose(%q): %d; want -1", text, got)
		}
	}
}

func TestQuotedPHPDocTypes(t *testing.T) {
	for _, c := range []struct{ doc, key string }{
		{`array{'first name': int}`, `first name`},
		{`array{'a}b': int}`, `a}b`},
		{`array{"a>b,(c)[d]": int}`, `a>b,(c)[d]`},
		{`array{'a\',b:c}': int}`, `a\',b:c}`},
	} {
		got := FromDoc(c.doc, nil)
		value, ok := got.ShapeKey(c.key)
		if !got.IsSealedShape() || !ok || !value.Equal(Int) {
			t.Errorf("%q: got %s; missing key %q", c.doc, got.ShapeString(), c.key)
		}
	}
	if got := FromDoc(`'a\'|b'|int`, nil); !got.Equal(Union(String, Int)) {
		t.Errorf("escaped union literal: %s", got)
	}
	if got := CallableReturn(FromDoc(`callable('a)b'): array{'x}y': int}`, nil)); !got.IsSealedShape() {
		t.Errorf("quoted callable parameter delimiter: %s", got.ShapeString())
	}
	for _, doc := range []string{`'`, `'unfinished`, `'unfinished\`, `'done'garbage`} {
		if got := FromDoc(doc, nil); !got.IsUnknown() {
			t.Errorf("unfinished or invalid literal %q: %s", doc, got)
		}
	}
	if got := FromDoc(`array{'unfinished}`, nil); got.HasShape() {
		t.Errorf("unfinished key claims shape: %s", got.ShapeString())
	}
}

func TestQuotedConditionalType(t *testing.T) {
	for _, c := range []struct{ doc, target, then string }{
		{`($x is 'a ? b' ? int : string)`, `'a ? b'`, `int`},
		{`($x is "a ? b" ? int : string)`, `"a ? b"`, `int`},
		{`($x is int ? 'a\' : b' : string)`, `int`, `string`},
	} {
		got, ok := ParseCond(c.doc, nil, nil)
		if !ok || got.Param != "x" || got.Target != c.target || got.Then.Type != c.then || got.Else.Type != "string" {
			t.Errorf("quoted conditional %q: %v, %v", c.doc, got, ok)
		}
		if got := FromDoc(c.doc, nil); !got.Has("string") {
			t.Errorf("quoted conditional branches %q: %s", c.doc, got)
		}
	}
	if _, ok := ParseCond(`($x is 'unfinished ? int : string)`, nil, nil); ok {
		t.Error("unfinished quote claims conditional")
	}
	if got := topSep(`$x is 'unfinished ? int : string`, '?', 0); got != -1 {
		t.Errorf("unfinished quote claims separator: %d", got)
	}
	for _, doc := range []string{`$x is 'a\' ? b' ? int : string`, `$x is "a\" ? b" ? int : string`} {
		if got, want := topSep(doc, '?', 0), strings.LastIndex(doc, "?"); got != want {
			t.Errorf("quoted conditional separator %q: %d; want %d", doc, got, want)
		}
		if got := FromDoc("("+doc+")", nil); !got.Equal(Union(Int, String)) {
			t.Errorf("quoted conditional union %q: %s", doc, got)
		}
	}
}

func BenchmarkFromDocQuoted(b *testing.B) {
	for name, doc := range map[string]string{
		"shape":    `array{'first name': int, 'a\',b:c}': array{"x>y": string}}`,
		"callable": `callable('a)b'): array{'x}y': int}`,
		"union":    `'a\'|b'|"c\"&d"|int`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				FromDoc(doc, nil)
			}
		})
	}
}
