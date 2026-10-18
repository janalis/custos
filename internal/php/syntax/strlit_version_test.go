package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestQuotedEscapeVersions(t *testing.T) {
	for _, test := range []struct {
		raw     string
		version phpversion.Version
		want    string
	}{
		{`"name\e"`, phpversion.PHP53, `name\e`},
		{`"name\e"`, phpversion.PHP54, "name\x1b"},
		{`"name\u{41}"`, phpversion.PHP56, `name\u{41}`},
		{`"name\u{41}"`, phpversion.PHP70, "nameA"},
		{`"name\u{41}"`, 0, "nameA"},
		{`'name\u{41}\e'`, phpversion.PHP85, `name\u{41}\e`},
		{`B"name\x41"`, phpversion.PHP53, "nameA"},
	} {
		got, ok := StringLiteralValueForVersion(test.raw, test.version)
		if !ok || got != test.want {
			t.Fatalf("%s PHP %v: %q/%v, want %q", test.raw, test.version, got, ok, test.want)
		}
	}
}

func TestParsedStringVersionAndHeredoc(t *testing.T) {
	for _, test := range []struct {
		source  string
		version phpversion.Version
		want    string
	}{
		{`"\u{30}"`, phpversion.PHP53, `\u{30}`},
		{`"\u{30}"`, phpversion.PHP70, "0"},
		{"<<<TEXT\n\\u{30}\nTEXT", phpversion.PHP53, `\u{30}`},
		{"<<<TEXT\n\\u{30}\nTEXT", phpversion.PHP70, "0"},
		{"<<<'TEXT'\n\\u{30}\nTEXT", phpversion.PHP70, `\u{30}`},
	} {
		file := Parse("literal.php", []byte("<?php "+test.source+";"), Options{Version: test.version})
		if len(file.Errors) != 0 {
			t.Fatal(file.Errors)
		}
		literal := file.Stmts[0].(*ExprStmt).Expr.(*Literal)
		value, ok := LiteralStringValue(literal)
		truth, known := constantTruth(literal)
		if !ok || value != test.want || !known || truth != (test.want != "" && test.want != "0") {
			t.Fatalf("PHP %v: %q/%v truth %v/%v, want %q", test.version, value, ok, truth, known, test.want)
		}
		if SourceVersion(literal) != test.version {
			t.Fatal("lost parsed version")
		}
	}
	if _, ok := LiteralStringValue(nil); ok {
		t.Fatal("nil string")
	}
	if _, ok := LiteralStringValue(&Literal{LitKind: LitInt, Raw: "1"}); ok {
		t.Fatal("number became text")
	}
	if SourceVersion((*Literal)(nil)) != phpversion.Max {
		t.Fatal("typed nil version")
	}
	if SourceVersion(nil) != phpversion.Max {
		t.Fatal("detached node version")
	}
	file := &File{}
	if SourceVersion(file) != phpversion.Max {
		t.Fatal("unspecified file version")
	}
	if value, ok := LiteralStringValue(&Literal{LitKind: LitString, Raw: `"\u{41}"`}); !ok || value != "A" {
		t.Fatal("detached literal")
	}
}

func BenchmarkParsedStringValue(b *testing.B) {
	file := Parse("literal.php", []byte(`<?php "ordinary text";`), Options{})
	literal := file.Stmts[0].(*ExprStmt).Expr.(*Literal)
	b.ReportAllocs()
	for b.Loop() {
		LiteralStringValue(literal)
	}
}
