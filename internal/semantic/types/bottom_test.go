package types

import "testing"

func TestBottomAndUnknown(t *testing.T) {
	bottom := Null.Without("null")
	if !bottom.IsBottom() || !bottom.IsUnknown() || bottom.Equal(Unknown) {
		t.Fatal("impossible narrowing must remain unusable as evidence but distinct from unknown")
	}
	for _, typ := range []Type{Unknown, Of(), Of(""), Int, Of("never")} {
		if typ.IsBottom() {
			t.Fatalf("%s is not a narrowing bottom", typ)
		}
	}
	if !Union(bottom, Int).Equal(Int) || !Union(Int, bottom).Equal(Int) || !Union(bottom, bottom).IsBottom() {
		t.Fatal("bottom must be neutral in joins")
	}
	if !Union(bottom, Unknown).Equal(Unknown) || !Union(Unknown, Int).Equal(Unknown) || !Union().Equal(Unknown) {
		t.Fatal("missing information must invalidate a join")
	}
	if !Ints(IntRange{Lo: 1, Hi: 1}).IntsWithout(1).IsBottom() {
		t.Fatal("exhausted value ranges must be bottom")
	}
}

func BenchmarkBottomUnion(b *testing.B) {
	bottom := Null.Without("null")
	b.ReportAllocs()
	for b.Loop() {
		Union(bottom, String)
	}
}

func TestArrayElementMetadata(t *testing.T) {
	generic := FromDoc(`Box<int>`, nil)
	array := Of(`\Box[]`).WithElem(generic)
	if !array.Elem().TypeArgs(`\Box`)[0].Equal(Int) || array.IsUnknown() {
		t.Fatal("lost generic element metadata")
	}
	for _, typ := range []Type{array, array.WithNonEmpty(true).WithArrayKey(String), Of(`\Box[][]`).WithElem(array), Union(array, Null)} {
		if got := FromDoc(typ.DocString(), nil); got.Elem().DocString() != typ.Elem().DocString() {
			t.Errorf("round trip %s: %s", typ.DocString(), got.Elem().DocString())
		}
		if typ.ShapeString() == "" {
			t.Fatal("missing debug representation")
		}
	}
	u := Union(array, Of("string[]"))
	if !u.Elem().Has("string") || len(u.Elem().TypeArgs(`\Box`)) != 1 {
		t.Fatal("union lost element metadata")
	}
	if !u.Without(`\Box[]`).Elem().Equal(String) {
		t.Fatal("removed array atom kept its elements")
	}
	if !Union(array, Array).Elem().IsUnknown() {
		t.Fatal("plain array acquired definite element metadata")
	}
	if !array.WithoutArrayInfo().Elem().Equal(Of(`\Box`)) || len(array.WithoutArrayInfo().Elem().TypeArgs(`\Box`)) != 0 {
		t.Fatal("metadata was not cleared")
	}
	if !Int.WithElem(generic).Equal(Int) {
		t.Fatal("non-array acquired elements")
	}
	relative := FromDoc("array<Box<static>>", nil)
	if !relative.HasRelative() {
		t.Fatal("relative generic element was not found")
	}
	if got := relative.BindRelative("Base", "Root", "Child").Elem().TypeArgs(`\Box`); len(got) != 1 || got[0].String() != `\Child` {
		t.Fatalf("relative elements: %v", got)
	}
	shaped := Of("int[]", "string[]", "null").WithShape([]ShapeKey{{Name: "value", Type: Int}}, true)
	if got := FromDoc(shaped.DocString(), nil); !got.Has("null") || !got.IsSealedShape() {
		t.Fatal("shape serialization lost keys or non-array members")
	}
}
