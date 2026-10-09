package types

import "testing"

func TestArrayKeyTemplates(t *testing.T) {
	for _, marker := range []string{`\~~K`, `\~K`} {
		for _, doc := range []string{"array<" + marker + ",int>", "non-empty-array<" + marker + ",mixed>", "array<string,array<" + marker + ",int>>"} {
			typ := FromDoc(doc, nil)
			keyed := typ
			if typ.ArrayKeyTemplate() == "" {
				keyed = typ.Elem()
			}
			if keyed.ArrayKeyTemplate() != marker || !keyed.ArrayKey().IsUnknown() {
				t.Fatalf("%s: marker %q, runtime key %s", doc, keyed.ArrayKeyTemplate(), keyed.ArrayKey())
			}
			if back := FromDoc(typ.DocString(), nil); back.DocString() != typ.DocString() {
				t.Fatalf("roundtrip %s => %s", typ.DocString(), back.DocString())
			}
		}
	}
	a := FromDoc(`array<\~~K,int>`, nil)
	b := FromDoc(`array<\~~K,string>`, nil)
	if Union(a, b).ArrayKeyTemplate() != `\~~K` || Union(a, Null).ArrayKeyTemplate() != `\~~K` {
		t.Fatal("compatible union lost key template")
	}
	for _, typ := range []Type{
		Unknown, Union(a, FromDoc(`array<\~~J,int>`, nil)), Union(a, FromDoc("array<string,int>", nil)),
		a.WithArrayKey(Int), a.WithArrayKey(Unknown), a.WithShape(nil, true), a.WithoutArrayInfo(),
		FromDoc(`array<\~~K|int,string>`, nil), FromDoc(`array<\~~K[],string>`, nil), FromDoc(`array<T,string>`, nil),
	} {
		if typ.ArrayKeyTemplate() != "" {
			t.Fatalf("stale template on %s", typ.DocString())
		}
	}
	if a.WithNonEmpty(true).WithoutShape().ArrayKeyTemplate() != `\~~K` || a.ArrayKeyTemplate() != `\~~K` {
		t.Fatal("immutable metadata change lost template")
	}
	if got := FromDoc(a.DocString(), func(string) string { return "=string" }); !got.ArrayKey().Equal(String) || got.ArrayKeyTemplate() != "" {
		t.Fatalf("resolved key %s", got.DocString())
	}
}

func TestCompoundArrayKeyTemplates(t *testing.T) {
	for _, marker := range []string{`\~K`, `\~~K`} {
		for _, fixed := range []string{"int", "string", "int|string"} {
			pattern := FromDoc(marker+"|"+fixed, nil)
			for _, doc := range []string{"array<" + marker + "|" + fixed + ",int>", "non-empty-array<" + marker + "|" + fixed + ",mixed>", "array<string,array<" + marker + "|" + fixed + ",int>>"} {
				typ := FromDoc(doc, nil)
				keyed := typ
				if typ.ArrayKeyPattern().IsUnknown() {
					keyed = typ.Elem()
				}
				if !keyed.ArrayKeyPattern().Equal(pattern) || !keyed.ArrayKey().IsUnknown() || keyed.ArrayKeyTemplate() != "" {
					t.Fatalf("%s: pattern %s, runtime %s", doc, keyed.ArrayKeyPattern(), keyed.ArrayKey())
				}
				if back := FromDoc(typ.DocString(), nil); back.DocString() != typ.DocString() {
					t.Fatalf("roundtrip %s => %s", typ.DocString(), back.DocString())
				}
			}
		}
	}
	a := FromDoc(`array<\~~K|int,int>`, nil)
	b := FromDoc(`array<int|\~~K,string>`, nil)
	if !Union(a, b).ArrayKeyPattern().Equal(a.ArrayKeyPattern()) || !Union(a, Null).ArrayKeyPattern().Equal(a.ArrayKeyPattern()) {
		t.Fatal("compatible union lost pattern")
	}
	for _, typ := range []Type{
		Unknown, Union(a, FromDoc(`array<\~~K|string,int>`, nil)), Union(a, FromDoc(`array<\~~J|int,int>`, nil)),
		Union(a, FromDoc(`array<int,int>`, nil)), a.WithArrayKey(Int), a.WithArrayKey(Unknown),
		a.WithShape(nil, true), a.WithShape(make([]ShapeKey, MaxShapeKeys+1), true), a.WithoutArrayInfo(),
		FromDoc(`array<\~~K|\~~J,int>`, nil), FromDoc(`array<\~~K|float,int>`, nil),
		FromDoc(`array<\~~K[]|int,int>`, nil), FromDoc(`array<\~~K<int>|int,int>`, nil),
		FromDoc(`array<(\~~K&\Foo)|int,int>`, nil), FromDoc(`array<T|int,int>`, nil),
	} {
		if !typ.ArrayKeyPattern().IsUnknown() {
			t.Fatalf("stale pattern on %s", typ.DocString())
		}
	}
	if !a.WithNonEmpty(true).WithoutShape().ArrayKeyPattern().Equal(a.ArrayKeyPattern()) || a.ArrayKeyPattern().IsUnknown() {
		t.Fatal("immutable change lost pattern")
	}
	resolved := FromDoc(a.DocString(), func(string) string { return "=string" })
	if !resolved.ArrayKey().Equal(Of("int", "string")) || !resolved.ArrayKeyPattern().IsUnknown() {
		t.Fatalf("resolved keys %s", resolved.DocString())
	}
}

func BenchmarkCompoundArrayKeyTemplates(b *testing.B) {
	for _, doc := range []string{`array<\~~K|int,int>`, `array<string,array<\~K|string,int>>`} {
		b.Run(doc, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				typ := FromDoc(doc, nil)
				_ = typ.ArrayKeyPattern()
				_ = typ.DocString()
			}
		})
	}
}
