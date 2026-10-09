package types

import (
	"slices"
	"testing"
)

func TestPHPDocMetadataIntersection(t *testing.T) {
	for _, doc := range []string{"?(Reader&Writer)", "(T is int ? Reader&Writer : null)", "(T is int ? Reader&Writer : Reader&Writer)"} {
		got := FromDoc(doc, nil)
		if !slices.Equal(got.Intersection(), []string{`\Reader`, `\Writer`}) {
			t.Errorf("%s: intersection %v", doc, got.Intersection())
		}
	}
	if got := FromDoc("(T is int ? Reader&Writer : Reader)", nil); got.Intersection() != nil {
		t.Errorf("alternative class retains intersection: %v", got.Intersection())
	}
}

func TestPHPDocMetadataShapes(t *testing.T) {
	for _, doc := range []string{"?array{id: int}", "(T is int ? array{id: int} : null)"} {
		got := FromDoc(doc, nil)
		id, ok := got.ShapeKey("id")
		if !got.Has("null") || !got.IsSealedShape() || !ok || !id.Equal(Int) {
			t.Errorf("%s: shape %s", doc, got.ShapeString())
		}
	}
	got := FromDoc("(T is int ? array{id: int, name: string} : array{id: string})", nil)
	id, ok := got.ShapeKey("id")
	if !ok || !id.Equal(Union(Int, String)) || !got.IsSealedShape() {
		t.Errorf("merged shape: %s", got.ShapeString())
	}
	keys := got.ShapeKeys()
	if len(keys) != 2 || !keys[1].Optional || keys[1].Name != "name" {
		t.Errorf("merged optional key: %v", keys)
	}
	if got := FromDoc("(T is int ? array{id: int} : array)", nil); got.HasShape() {
		t.Errorf("unshaped branch retains shape: %s", got.ShapeString())
	}
}

func TestPHPDocMetadataGenericsAndCallables(t *testing.T) {
	for _, doc := range []string{"?Collection<string>", "(T is int ? Collection<string> : null)", "(T is int ? Collection<string> : Collection<string>)"} {
		got := FromDoc(doc, nil).TypeArgs(`\Collection`)
		if len(got) != 1 || !got[0].Equal(String) {
			t.Errorf("%s: arguments %v", doc, got)
		}
	}
	for _, doc := range []string{"(T is int ? Collection<int> : Collection<string>)", "(T is int ? Collection<int> : Collection)"} {
		if got := FromDoc(doc, nil).TypeArgs(`\Collection`); got != nil {
			t.Errorf("%s: conflicting arguments %v", doc, got)
		}
	}
	for _, doc := range []string{"?callable(): array{id: int}", "(T is int ? callable(): array{id: int} : null)", "(T is int ? callable(): array{id: int} : callable(): array{id: int})"} {
		ret := CallableReturn(FromDoc(doc, nil))
		id, ok := ret.ShapeKey("id")
		if !ok || !id.Equal(Int) {
			t.Errorf("%s: callable return %s", doc, ret.ShapeString())
		}
	}
	for _, doc := range []string{"(T is int ? callable(): int : callable(): string)", "(T is int ? callable(): int : callable)"} {
		if got := CallableReturn(FromDoc(doc, nil)); !got.IsUnknown() {
			t.Errorf("%s: conflicting callable return %s", doc, got)
		}
	}
}

func TestPHPDocMetadataUnknown(t *testing.T) {
	for _, doc := range []string{"?{invalid}", "(T is int ? int : {invalid})", "(T is int ? {invalid} : int)"} {
		if got := FromDoc(doc, nil); !got.IsUnknown() {
			t.Errorf("%s: partially known type %s", doc, got)
		}
	}
}
