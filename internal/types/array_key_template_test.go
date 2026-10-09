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
