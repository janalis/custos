package types

import "testing"

// An empty array joins array members that admit one: the union reads their
// elements and is no longer known non-empty. A sealed shape makes its keys
// optional instead, and an empty array alone stays one.
func TestUnionAbsorbsEmptyArrays(t *testing.T) {
	empty := FromDoc("array{}", nil)
	for _, tc := range []struct {
		members []Type
		want    string
		elem    string
	}{
		{[]Type{FromDoc(`\Foo[]`, nil), empty}, `\Foo[]`, `\Foo`},
		{[]Type{empty, FromDoc(`list<\Foo>`, nil), Null}, `\Foo[]|null`, `\Foo`},
		{[]Type{FromDoc(`non-empty-list<\Foo>`, nil), empty}, `\Foo[]`, `\Foo`},
		{[]Type{FromDoc(`array{a: int, ...}`, nil), empty}, `array{a?: int, ...}`, "?unknown"},
		{[]Type{FromDoc(`array`, nil), empty}, "array", "?unknown"},
		{[]Type{FromDoc(`array{a: int}`, nil), empty}, `array{a?: int}`, "?unknown"},
		{[]Type{empty, Null}, "array|null{}", "?unknown"},
		{[]Type{empty, empty}, "array{}", "?unknown"},
	} {
		u := Union(tc.members...)
		if u.ShapeString() != tc.want || u.Elem().String() != tc.elem || u.IsNonEmptyArray() {
			t.Errorf("%v: got %s, elements %s, non-empty %v", tc.members, u.ShapeString(), u.Elem(), u.IsNonEmptyArray())
		}
	}
}

func TestArrayCardinality(t *testing.T) {
	for _, tc := range []struct {
		doc              string
		minimum, maximum int
		known            bool
	}{
		{"array{}", 0, 0, true},
		{"array{a: int, b?: string}", 1, 2, true},
		{"array{a:int,b:string}", 2, 2, true},
		{"array", 0, 0, false},
		{"array{}|null", 0, 0, false},
		{"int", 0, 0, false},
	} {
		minimum, maximum, known := FromDoc(tc.doc, nil).ArrayCardinality()
		if minimum != tc.minimum || maximum != tc.maximum || known != tc.known {
			t.Fatalf("%s: %d %d %v", tc.doc, minimum, maximum, known)
		}
	}
}
