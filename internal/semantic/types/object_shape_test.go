package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestObjectShapeRoundTrip(t *testing.T) {
	for _, c := range []struct{ in, doc string }{
		{"object{a: int, b?: string}", "object{a:int,b?:string}"},
		{"object{a: int, ...}", "object{a:int,...}"},
		{"object{}", "object{}"},
		{"object{...}", "object{...}"},
		{"object{'a-b': array{c: int}}", "object{'a-b':array{c:int}}"},
		{"object{a: int}|null", "null|object{a:int}"},
	} {
		got := FromDoc(c.in, nil)
		if !got.Has("object") {
			t.Fatalf("%s: atoms %v", c.in, got.atoms)
		}
		if s := got.DocString(); s != c.doc {
			t.Errorf("%s: doc = %s, want %s", c.in, s, c.doc)
		}
		if again := FromDoc(got.DocString(), nil).DocString(); again != c.doc {
			t.Errorf("%s: round trip = %s", c.in, again)
		}
		if !reflect.DeepEqual(roundTrip(t, got), got) {
			t.Errorf("%s: wire round trip differs", c.in)
		}
	}
}

func TestObjectShapePlainFallback(t *testing.T) {
	many := make([]string, MaxShapeKeys+1)
	for i := range many {
		many[i] = "k" + strings.Repeat("x", i) + ": int"
	}
	for _, in := range []string{"object", "object{0: int}", "object{int}", "object{a: int, a: string}", "object{a: int", "object{" + strings.Join(many, ",") + "}"} {
		if got := FromDoc(in, nil); got.DocString() != "object" {
			t.Errorf("%s = %s, want object", in, got.DocString())
		}
	}
}

func TestObjectShapeInUnions(t *testing.T) {
	a := FromDoc("object{a: int}", nil)
	if s := Union(a, Null).DocString(); s != "null|object{a:int}" {
		t.Errorf("with null = %s", s)
	}
	if s := Union(a, FromDoc("object{a: int}", nil)).DocString(); s != "object{a:int}" {
		t.Errorf("same shape = %s", s)
	}
	for _, other := range []string{"object", "object{a: string}", "object{a?: int}", "object{a: int, ...}", "object{a: int, b: int}", "object{b: int}"} {
		if s := Union(a, FromDoc(other, nil)).DocString(); s != "object" {
			t.Errorf("with %s = %s, want object", other, s)
		}
	}
	if s := FromDoc("object{a: int}|object{b: int}", nil).DocString(); s != "object" {
		t.Errorf("doc union = %s", s)
	}
}

func TestObjectShapeShapeString(t *testing.T) {
	if s := FromDoc("object{a: int}", nil).ShapeString(); !strings.Contains(s, "object{a:int}") {
		t.Errorf("ShapeString = %s", s)
	}
}

func TestObjectShapeBindRelative(t *testing.T) {
	in := FromDoc("object{a: int}|callable(self): self", nil)
	got := in.BindRelative(`\Foo`, `\Foo`, "")
	if s := got.DocString(); !strings.Contains(s, "object{a:int}") {
		t.Errorf("bound = %s", s)
	}
}
