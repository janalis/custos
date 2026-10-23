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
