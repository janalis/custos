package types

import (
	"strings"
	"testing"
)

func TestArrayKeyMetadata(t *testing.T) {
	for doc, want := range map[string]string{
		"array<string,int>": "string", "array<int,int>": "int", "array<array-key,int>": "int|string",
		"list<int>": "int", "non-empty-list<int>": "int", "non-empty-list": "int", "list": "int", "array<int>": "?unknown", "non-empty-array": "?unknown",
		"array<T,int>": "?unknown", "array<string,int": "?unknown", "array<int,string>|array<string,int>": "int|string",
		"array<int,string>|string[]": "?unknown", "?array<string,int>": "string",
		"array{a:int,2:string}": "int|string", "array{}": "?unknown", "array{a:int,...}": "?unknown",
		"array{...<string,mixed>}": "string", "non-empty-array{...<string,mixed>}": "string", "array{...<int,string>}": "?unknown",
	} {
		typ := FromDoc(doc, nil)
		if got := typ.ArrayKey().String(); got != want {
			t.Errorf("%s: %s, want %s", doc, got, want)
		}
		back := FromDoc(typ.DocString(), nil)
		if !back.Equal(typ) || !back.ArrayKey().Equal(typ.ArrayKey()) {
			t.Errorf("%s roundtrip %s => %s", doc, typ.DocString(), back)
		}
	}
	nested := FromDoc("array<string,list<int>>", nil)
	if !nested.Elem().ArrayKey().Equal(Int) {
		t.Fatal("nested key metadata lost")
	}
	if !FromDoc(nested.DocString(), nil).Elem().ArrayKey().Equal(Int) {
		t.Fatal("nested roundtrip lost keys")
	}
	for _, key := range []Type{Unknown, Float, Of("int", "float")} {
		if !Array.WithArrayKey(key).ArrayKey().IsUnknown() {
			t.Fatal("unsupported domain retained")
		}
	}
	if !String.WithArrayKey(Int).ArrayKey().IsUnknown() {
		t.Fatal("non-array retained domain")
	}
	for _, key := range []Type{Int, String, Of("int", "string")} {
		typ := Array.WithArrayKey(key)
		if !FromDoc(typ.DocString(), nil).ArrayKey().Equal(key) {
			t.Fatal("key-only roundtrip")
		}
	}
	if !Array.WithArrayKey(String).WithShape(make([]ShapeKey, MaxShapeKeys+1), true).ArrayKey().IsUnknown() {
		t.Fatal("oversized replacement shape retained stale domain")
	}
	a := FromDoc("array{a:int}", nil)
	if !a.WithoutShape().ArrayKey().Equal(String) || !a.HasShape() {
		t.Fatal("dropping shape lost keys or mutated source")
	}
	if !a.WithoutArrayInfo().ArrayKey().IsUnknown() {
		t.Fatal("array info removal retained keys")
	}
	if !Array.WithArrayKey(Int).WithNonEmpty(true).ArrayKey().Equal(Int) {
		t.Fatal("emptiness lost keys")
	}
}

func TestAnonymousDocRoundTrip(t *testing.T) {
	marker := `\@anonymous:` + strings.Repeat("a", 64) + ":17"
	for _, doc := range []string{marker, marker + "[]", "array<string," + marker + ">", "Collection<" + marker + ">"} {
		typ := FromDoc(doc, nil)
		if typ.IsUnknown() || !FromDoc(typ.DocString(), nil).Equal(typ) {
			t.Errorf("anonymous roundtrip %s", doc)
		}
	}
	for _, bad := range []string{`\@anonymous:abc:1`, `\@anonymous:` + strings.Repeat("A", 64) + ":1", marker + ":1", `\@anonymous:` + strings.Repeat("a", 64) + ":4294967296"} {
		if !FromDoc(bad, nil).IsUnknown() {
			t.Errorf("invalid marker accepted: %s", bad)
		}
	}
}

func BenchmarkArrayKeyMetadata(b *testing.B) {
	for _, doc := range []string{"array<string,int>", "array<string,list<int>>", "array{a:int}|array<int,string>"} {
		b.Run(doc, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				typ := FromDoc(doc, nil)
				_ = typ.ArrayKey()
				_ = typ.DocString()
			}
		})
	}
}
