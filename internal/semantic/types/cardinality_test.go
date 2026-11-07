package types

import "testing"

func TestIndependentArrayCardinality(t *testing.T) {
	two := Array.WithArrayCardinality(2, 2)
	for _, tc := range []struct {
		value     Type
		low, high int
		known     bool
	}{
		{two, 2, 2, true},
		{Union(two, Array.WithArrayCardinality(4, 5)), 2, 5, true},
		{Union(two, FromDoc("array{}", nil)), 0, 2, true},
		{Union(FromDoc("array{a:int}", nil), FromDoc("array{b:string}", nil)), 1, 1, true},
		{Union(Bottom, two), 2, 2, true},
		{Union(two, Array), 0, 0, false},
		{two.WithNonEmpty(false), 0, 2, true},
		{two.WithoutShape(), 0, 0, false},
		{two.WithoutArrayInfo(), 0, 0, false},
		{two.WithShape([]ShapeKey{{Name: "a", Type: Int}}, true), 1, 1, true},
		{two.WithArrayCardinality(-1, 3), 0, 0, false},
		{two.WithArrayCardinality(3, 1), 0, 0, false},
		{Int.WithArrayCardinality(1, 1), 0, 0, false},
	} {
		low, high, known := tc.value.ArrayCardinality()
		if low != tc.low || high != tc.high || known != tc.known {
			t.Fatalf("%s: %d %d %v", tc.value.ShapeString(), low, high, known)
		}
	}
	shape := FromDoc("array{a:int}", nil).WithArrayCardinality(1, 1)
	changed := shape.MapShape(func(k ShapeKey) ShapeKey { k.Optional = true; return k })
	if low, _, _ := changed.ArrayCardinality(); low != 0 {
		t.Fatal("changed optionality retained exact count")
	}
	unchanged := shape.MapShape(func(k ShapeKey) ShapeKey { k.Type = String; return k })
	if low, high, _ := unchanged.ArrayCardinality(); low != 1 || high != 1 {
		t.Fatal("type substitution lost count")
	}
	if !two.IsNonEmptyArray() {
		t.Fatal("count did not imply nonempty")
	}
}

func BenchmarkCardinalityUnion(b *testing.B) {
	a, c := Array.WithArrayCardinality(2, 3), Array.WithArrayCardinality(5, 6)
	b.ReportAllocs()
	for b.Loop() {
		Union(a, c)
	}
}

func TestZeroCountArrayAlternatives(t *testing.T) {
	empty := Array.WithArrayCardinality(0, 0)
	other := Of("string[]").WithShape([]ShapeKey{{Name: "0", Type: String}}, true)
	result := Union(empty, other, empty)
	if result.Has("array") || !result.Elem().Equal(String) {
		t.Fatalf("empty alternative widened elements: %s", result.ShapeString())
	}
	if low, high, known := result.ArrayCardinality(); !known || low != 0 || high != 1 {
		t.Fatalf("union size: %d %d %v", low, high, known)
	}
	if _, _, known := Union(Unknown, empty).ArrayCardinality(); known {
		t.Fatal("unknown union retained proof")
	}
}
