package syntax

import (
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestDecodeLiteralString(t *testing.T) {
	for _, tc := range []struct {
		raw, value string
		sources    []string
	}{
		{`'a\'b'`, "a'b", []string{"a", `\'`, "b"}},
		{`b"\x41\101\u{1f600}\n"`, "AA😀\n", []string{`\x41`, `\101`, `\u{1f600}`, `\u{1f600}`, `\u{1f600}`, `\u{1f600}`, `\n`}},
		{`'\n'`, `\n`, []string{`\`, "n"}},
		{"<<<'END'\n  ab\n  cd\n  END", "ab\ncd", []string{"a", "b", "\n", "c", "d"}},
		{"<<<END\r\n  a\\n\r\n  END", "a\n", []string{"a", `\n`}},
		{"<<<'END'\nEND", "", nil},
		{"<<<END\nx\nEND", "x", []string{"x"}},
	} {
		f := Parse("test.php", []byte("<?php $s = "+tc.raw+";"), Options{Version: phpversion.Max})
		var lit *Literal
		InspectFile(f, func(n Node) bool {
			if l, ok := n.(*Literal); ok {
				lit = l
			}
			return true
		})
		got, ok := DecodeLiteralString(lit)
		if !ok || got.Value != tc.value || len(got.Sources) != len(tc.sources) {
			t.Fatalf("%q: %+v, %v", tc.raw, got, ok)
		}
		for i, span := range got.Sources {
			if string(f.Src[span.Start:span.End]) != tc.sources[i] {
				t.Fatalf("%q byte %d: %q", tc.raw, i, f.Src[span.Start:span.End])
			}
		}
	}
	if _, ok := DecodeLiteralString(nil); ok {
		t.Fatal("accepted nil")
	}
}

func TestStringEscapeMapping(t *testing.T) {
	for _, raw := range []string{`"\xZ\q\0\uZ\"\\\$"`, `'abc'`, `"é"`, `""`} {
		l := &Literal{LitKind: LitString, Raw: raw}
		got, ok := DecodeLiteralString(l)
		if !ok || len(got.Sources) != len(got.Value) {
			t.Fatalf("%s: %+v %v", raw, got, ok)
		}
	}
}

func BenchmarkDecodeLiteralString(b *testing.B) {
	l := &Literal{LitKind: LitString, Raw: `"` + strings.Repeat(`ab\x41\u{1f600}`, 100) + `"`}
	b.ReportAllocs()
	for b.Loop() {
		DecodeLiteralString(l)
	}
}

func TestStringMappingVersionAndSurrogate(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		version phpversion.Version
	}{
		{`"\u{d800}"`, phpversion.PHP84},
		{`"\e\u{41}"`, phpversion.PHP53},
	} {
		f := Parse("t.php", []byte("<?php "+tc.raw+";"), Options{Version: tc.version})
		literal := f.Stmts[0].(*ExprStmt).Expr.(*Literal)
		got, ok := DecodeLiteralString(literal)
		want, _ := LiteralStringValue(literal)
		if !ok || got.Value != want || len(got.Sources) != len(want) {
			t.Fatalf("%s: %+v", tc.raw, got)
		}
	}
}

func FuzzDecodeLiteralString(f *testing.F) {
	for _, raw := range []string{`'plain'`, `"\x41\u{1f600}"`, "<<<'END'\n  a\n  END", `"\u{d800}"`, ""} {
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		literal := &Literal{LitKind: LitString, Raw: raw}
		got, ok := DecodeLiteralString(literal)
		if !ok {
			return
		}
		if len(got.Sources) != len(got.Value) {
			t.Fatal("mapping length")
		}
		for _, span := range got.Sources {
			if span.Start >= span.End || int(span.End) > len(raw) {
				t.Fatal("mapping range")
			}
		}
	})
}
