package types

import "testing"

func TestAcceptsShapes(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	for _, c := range []struct {
		expected, actual string
		want             Verdict
	}{
		// A required key a sealed actual lacks.
		{`array{a: int}`, `array{b: int}`, No},
		{`array{a: int}`, `array{}`, No},
		{`array{a: int, b?: int}`, `array{b: int}`, No},
		// Open or optional actual keys may hold it.
		{`array{a: int, ...}`, `array{b: int, ...}`, Maybe},
		{`array{a: int}`, `array{b: int, ...}`, No}, // b is required and not listed
		{`array{a: int}`, `array{a?: string}`, Maybe},
		{`array{a?: int, ...}`, `array{b: int}`, Maybe},
		{`array{a?: int}`, `array{b: int}`, No}, // b is required and not listed
		// Values of keys required on both sides.
		{`array{a: int}`, `array{a: string}`, No},
		{`array{a: int}`, `array{a: int}`, Maybe},
		{`array{a: int|string}`, `array{a: int}`, Maybe},
		// A sealed expected shape lists every required key.
		{`array{a: int}`, `array{a: int, b: int}`, No},
		{`array{a: int}`, `array{a: int, b?: int}`, Maybe},
		{`array{a: int, ...}`, `array{a: int, b: int}`, Maybe},
		// Arrays without a shape.
		{`array{a: int}`, `list<int>`, No},
		{`array{a: int}`, `array<int, string>`, No},
		{`array{0: int}`, `list<int>`, Maybe},
		{`array{a?: int}`, `list<int>`, Maybe},
		{`array{a: int}`, `array<string, int>`, Maybe},
		{`array{a: int}`, `array`, Maybe},
	} {
		if got := Accepts(FromDoc(c.expected, nil), FromDoc(c.actual, nil), ctx); got != c.want {
			t.Errorf("Accepts(%s, %s) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

func TestAcceptsArrayKeys(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	for _, c := range []struct {
		expected, actual string
		want             Verdict
	}{
		{`array<string, int>`, `non-empty-array<int, int>`, No},
		{`array<int, int>`, `non-empty-array<string, int>`, Maybe}, // numeric strings are int keys
		{`non-empty-array<string, int>`, `non-empty-list<int>`, No},
		{`array<string, int>`, `array<int, int>`, Maybe},
		{`array<string, int>`, `non-empty-array<string, int>`, Maybe},
		{`array<int|string, int>`, `non-empty-array<int, int>`, Maybe},
		{`array<string, int>`, `non-empty-array`, Maybe},
		{`array<string, int>`, `array{}`, Yes},
	} {
		if got := Accepts(FromDoc(c.expected, nil), FromDoc(c.actual, nil), ctx); got != c.want {
			t.Errorf("Accepts(%s, %s) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

// A class constant in a positional entry is no `name: type` key.
func TestShapeConstantEntryIsPositional(t *testing.T) {
	got := FromDoc(`array{self::A|self::B, int}`, nil)
	keys := got.ShapeKeys()
	if len(keys) != 2 || keys[0].Name != "0" || keys[1].Name != "1" {
		t.Errorf("keys = %+v", keys)
	}
}

func TestAcceptsObjectShapes(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	for _, c := range []struct {
		expected, actual string
		want             Verdict
	}{
		{`object{a: int}`, `object{a: string}`, No},
		{`object{a: int}`, `object{a: int}`, Yes},
		{`object{a: int}`, `object{b: string}`, Yes},
		{`object{a?: int}`, `object{a: string}`, Yes},
		{`object{a: int}`, `object{a?: string}`, Yes},
		{`object{a: int}`, `object`, Yes},
		{`object`, `object{a: int}`, Yes},
	} {
		if got := Accepts(FromDoc(c.expected, nil), FromDoc(c.actual, nil), ctx); got != c.want {
			t.Errorf("Accepts(%s, %s) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

func TestWithoutClassTypeArgsInShapes(t *testing.T) {
	in := FromDoc(`array{a: \A<int>}`, nil)
	if got := in.WithoutClassTypeArgs().DocString(); got != `array{a:\A}` {
		t.Errorf("stripped = %s", got)
	}
	if got := in.DocString(); got != `array{a:\A<int>}` {
		t.Errorf("the original changed: %s", got)
	}
}
