package types

import "testing"

// ValueDocString keeps the int values and non-empty strings DocString
// drops, in a form FromDoc reads back with the same values; it is empty
// without such facts.
func TestValueDocString(t *testing.T) {
	for _, tc := range []struct{ doc, want string }{
		{"positive-int", "positive-int"},
		{"negative-int|null", "negative-int|null"},
		{"non-negative-int", "non-negative-int"},
		{"non-positive-int", "non-positive-int"},
		{"non-zero-int", "negative-int|positive-int"},
		{"int<1, 9>|Foo", `int<1, 9>|\Foo`},
		{"non-empty-string", "non-empty-string"},
		{"non-empty-string|int", "int|non-empty-string"},
		{"positive-int|string", "positive-int|string"},
		{"numeric-string|array{a: int}", "non-empty-string|array{a:int}"},
		{"int", ""},
		{"string|null", ""},
		{"'a'|'b'", ""},
		{"Foo", ""},
	} {
		typ := FromDoc(tc.doc, nil)
		got := typ.ValueDocString()
		if got != tc.want {
			t.Errorf("%s: got %q want %q", tc.doc, got, tc.want)
			continue
		}
		if got != "" && FromDoc(got, nil).ShapeString() != typ.ShapeString() {
			t.Errorf("%s: %s reads back as %s, not %s", tc.doc, got, FromDoc(got, nil).ShapeString(), typ.ShapeString())
		}
	}
}

// WithoutValues keeps the atoms and drops the values of ints and strings.
func TestWithoutValues(t *testing.T) {
	typ := Union(IntValue(3), StringLength(2), Null)
	if got := typ.WithoutValues().ShapeString(); got != "int|null|string" {
		t.Errorf("got %s", got)
	}
}
