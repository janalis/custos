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
