package types

import (
	"testing"

	"custos/internal/syntax"
)

func TestFromDoc(t *testing.T) {
	resolve := func(n string) string {
		if n == "Foo" {
			return `App\Foo`
		}
		if n[0] == '\\' {
			return n[1:]
		}
		return n
	}
	cases := map[string]string{
		"int|null":                       "int|null",
		"?string":                        "null|string",
		"Foo[]":                          `\App\Foo[]`,
		"array<int, Foo>":                `\App\Foo[]`,
		"list<string>":                   "string[]",
		"array{a: int, b?: string}":      "array",
		"\\Traversable&\\Countable":      `\Countable|\Traversable`,
		"positive-int|false":             "false|int",
		"class-string<Foo>":              "string",
		"callable(int): void":            "callable",
		"Collection<int, Foo>":           `\Collection`,
		"(int|string)[]":                 "int[]|string[]",
		"$this":                          "static",
		"'a'|'b'":                        "string",
		"integer|boolean|double":         "bool|float|int",
		"\\Generator<int, string, void>": `\Generator`,
		"array<string, string|false>":    "false[]|string[]",
		"array<string, array<string, mixed>|scalar|null>":      "bool[]|float[]|int[]|mixed[][]|null[]|string[]",
		"(TKey is null ? array<string, mixed> : array<mixed>)": "mixed[]",
		"($x is not null ? Foo : null)":                        `\App\Foo|null`,
		"(T is int ? int : (T is string ? string : false))":    "false|int|string",
	}
	for in, want := range cases {
		if got := FromDoc(in, resolve).String(); got != want {
			t.Errorf("%q: got %s want %s", in, got, want)
		}
	}
	if !FromDoc("", nil).IsUnknown() {
		t.Error("empty doc type must be unknown")
	}
	for _, in := range []string{"{array}", "self::TYPE_*", "foo-bar", `A\\B`, "1abc"} {
		if got := FromDoc(in, resolve); !got.IsUnknown() && got.String() != "int" {
			t.Errorf("%q: invalid class name typed as %s", in, got)
		}
	}
	if got := FromDoc("int|{array}", resolve).String(); got != "int" {
		t.Errorf("int|{array}: got %s", got)
	}
}

func TestTypeOps(t *testing.T) {
	a := Of("int", "null")
	if !a.Has("NULL") || !a.IsNullable() || a.Without("null").String() != "int" {
		t.Fatal(a)
	}
	if !Union(a, String).Equal(Of("string", "null", "int")) {
		t.Fatal("union")
	}
	if !Union(a, Unknown).IsUnknown() {
		t.Fatal("union with unknown must be unknown")
	}
	if Of(`\Foo[]`).Elem().String() != `\Foo` || !Of("int[]", "array").IsArrayLike() {
		t.Fatal("arrays")
	}
}

func TestFromNodeNoAliases(t *testing.T) {
	// In a native declaration `Double`/`integer` are class names, not aliases.
	if got := FromNode(&syntax.Name{Value: "Double"}, nil).String(); got != `\Double` {
		t.Errorf("Double: got %s", got)
	}
	if got := FromNode(&syntax.Name{Value: "int"}, nil).String(); got != "int" {
		t.Errorf("int: got %s", got)
	}
}

func TestDocShapes(t *testing.T) {
	cases := map[string]string{
		"array{path: string, line: int}":         "array{path: string, line: int}",
		"array{a: int, b?: string}":              "array{a: int, b?: string}",
		"array{'quoted key': int, 0: bool}":      "array{quoted key: int, 0: bool}",
		"list{int, string}":                      "array{0: int, 1: string}",
		"array{int, string}":                     "array{0: int, 1: string}",
		"array{a: int, ...}":                     "array{a: int, ...}",
		"array{}":                                "array{}",
		"array{a: array{b: float}}|null":         "array|null{a: array{b: float}}",
		"non-empty-array<string>":                "non-empty string[]",
		"non-empty-list":                         "non-empty array",
		"list<array{id: int}>":                   "array[]<array{id: int}>",
		"array{id: int}[]":                       "array[]<array{id: int}>",
		"array{a: int}|array{a: string, b: int}": "array{a: int|string, b?: int}",
		"array{a: int}|array<int>":               "array|int[]",
		"array{a: callable(int): void}":          "array{a: callable}",
	}
	for in, want := range cases {
		got := FromDoc(in, nil)
		if got.ShapeString() != want {
			t.Errorf("%s: got %s want %s", in, got.ShapeString(), want)
		}
		// The atom model is unchanged by shapes.
		if !got.Has("array") && !got.IsArrayLike() && !got.Has("null") {
			t.Errorf("%s: atoms %s lost their array member", in, got)
		}
		// DocString round-trips through FromDoc.
		if back := FromDoc(got.DocString(), nil); back.ShapeString() != got.ShapeString() {
			t.Errorf("%s: DocString %q parses back as %s", in, got.DocString(), back.ShapeString())
		}
	}
}

func TestShapeAtomsCompatible(t *testing.T) {
	s := FromDoc("array{a: int}", nil)
	if !s.IsArrayLike() || !s.Has("array") || s.String() != "array" || !s.Equal(Array) {
		t.Fatalf("shape must keep the plain array atom: %s", s.ShapeString())
	}
	if kt, ok := s.ShapeKey("a"); !ok || kt.String() != "int" {
		t.Fatalf("ShapeKey(a) = %s, %v", kt, ok)
	}
	if !s.IsNonEmptyArray() {
		t.Fatal("a shape with a required key is non-empty")
	}
	u := Union(s, Null)
	if !u.HasShape() || u.String() != "array|null" {
		t.Fatalf("union with null keeps the shape: %s", u.ShapeString())
	}
	if w := u.Without("null"); !w.HasShape() {
		t.Fatalf("Without keeps the shape: %s", w.ShapeString())
	}
	if w := u.Without("array"); w.HasShape() {
		t.Fatalf("no array member, no shape: %s", w.ShapeString())
	}
	if m := Union(s, Array); m.HasShape() {
		t.Fatalf("a plain array member drops the shape: %s", m.ShapeString())
	}
	keys := make([]ShapeKey, MaxShapeKeys+1)
	for i := range keys {
		keys[i] = ShapeKey{Name: string(rune('a'+i%26)) + string(rune('a'+i/26)), Type: Int}
	}
	big := Array.WithShape(keys, true)
	if big.HasShape() || !big.IsNonEmptyArray() {
		t.Fatalf("over the cap: no shape but still non-empty: %s", big.ShapeString())
	}
	if ne := Of("int[]").WithNonEmpty(true); !ne.IsNonEmptyArray() || Union(ne, Of("int[]")).IsNonEmptyArray() {
		t.Fatal("non-empty only survives unions of non-empty members")
	}
}

// BenchmarkUnion covers the hot path of variable typing: plain atom unions
// (no array facts) and unions of shaped arrays.
func BenchmarkUnion(b *testing.B) {
	plain := []Type{Of("int"), Of("null", "string"), Of(`\Foo`)}
	shaped := []Type{FromDoc("array{a: int, b: string}", nil), FromDoc("array{a: string, c?: bool}|null", nil)}
	b.Run("plain", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Union(plain...)
		}
	})
	b.Run("shaped", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = Union(shaped...)
		}
	})
}

func TestNestedAliasShapes(t *testing.T) {
	aliases := map[string]string{
		"Fn":   "array{name: string, paths: array<int, int>}",
		"Fns":  "array<string, Fn>",
		"Loop": "array{next: Loop}",
	}
	resolve := func(n string) string {
		if d, ok := aliases[n]; ok {
			return "=" + d
		}
		return n
	}
	cases := map[string]string{
		"Fns":               "array[]<array{name: string, paths: int[]}>",
		"array{Fn, string}": "array{0: array{name: string, paths: int[]}, 1: string}",
		"Loop":              "array{next: mixed}",
	}
	for in, want := range cases {
		if got := FromDoc(in, resolve).ShapeString(); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}
