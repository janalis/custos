package phpdoc

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseMethod(t *testing.T) {
	tests := []struct {
		text string
		want MethodSignature
	}{
		{"ping()", MethodSignature{Name: "ping"}},
		{" static\tself make() description", MethodSignature{Name: "make", Return: "self", Static: true}},
		{"void fill(array &$items, string &...$rest)", MethodSignature{Name: "fill", Return: "void", Params: []MethodParam{
			{Name: "items", Type: "array", ByRef: true}, {Name: "rest", Type: "string", ByRef: true, Variadic: true, Optional: true},
		}}},
		{"array<int, Foo> load(array{path: string, tags?: list<int>} $row, callable(string $x, int): bool $callback, $untyped)", MethodSignature{Name: "load", Return: "array<int, Foo>", Params: []MethodParam{
			{Name: "row", Type: "array{path: string, tags?: list<int>}"}, {Name: "callback", Type: "callable(string $x, int): bool"}, {Name: "untyped"},
		}}},
		{"(Foo & Bar)|null find((Foo & Bar)|null $item = null)", MethodSignature{Name: "find", Return: "(Foo & Bar)|null", Params: []MethodParam{
			{Name: "item", Type: "(Foo & Bar)|null", Optional: true, Default: "null"},
		}}},
		{"callable(string): int callback()", MethodSignature{Name: "callback", Return: "callable(string): int"}},
		{"callable(string) : int callback()", MethodSignature{Name: "callback", Return: "callable(string) : int"}},
		{"callable(string $x) callback()", MethodSignature{Name: "callback", Return: "callable(string $x)"}},
		{"Closure(string) callback()", MethodSignature{Name: "callback", Return: "Closure(string)"}},
		{`\Closure(string) callback()`, MethodSignature{Name: "callback", Return: `\Closure(string)`}},
		{"callable()", MethodSignature{Name: "callable"}},
		{"callable() description", MethodSignature{Name: "callable"}},
		{`void defaults(string $a = 'x, y', string $b = "x, \"y", array $c = ['key' => [1, 2]], mixed $d = Factory::make(1, 2)) trailing`, MethodSignature{Name: "defaults", Return: "void", Params: []MethodParam{
			{Name: "a", Type: "string", Optional: true, Default: `'x, y'`},
			{Name: "b", Type: "string", Optional: true, Default: `"x, \"y"`},
			{Name: "c", Type: "array", Optional: true, Default: "['key' => [1, 2]]"},
			{Name: "d", Type: "mixed", Optional: true, Default: "Factory::make(1, 2)"},
		}}},
		{"void café($é)", MethodSignature{Name: "café", Return: "void", Params: []MethodParam{{Name: "é"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, ok := ParseMethod(tt.text)
			if !ok || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseMethod = %+v, %v; want %+v", got, ok, tt.want)
			}
		})
	}
}

func TestParseMethodMalformed(t *testing.T) {
	for _, text := range []string{
		"", "noParens", "()", "int 123()", "int wrong-name()", "int work()junk",
		"int work(", "int work(int)", "int work($)", "int work($123)",
		"int work($x trailing)", "int work($x = )", "int work(...$x = [])",
		"int work(&&$x)", "int work(......$x)", "int work($x,)", "int work(, $x)",
		"int work(array<int] $x)", "int work($x = 'unclosed)", "int work($x = ]))",
		"int work($x = 'escaped\\')", "int work($x = (1, 2])", "int ]work()",
		strings.Repeat("x", maxMethodLen+1),
		"int work(" + strings.Repeat("[", maxMethodDepth) + "$x" + strings.Repeat("]", maxMethodDepth) + ")",
		"int work(" + strings.Repeat("$x,", maxMethodParams) + "$last)",
	} {
		t.Run(text[:min(len(text), 70)], func(t *testing.T) {
			if got, ok := ParseMethod(text); ok {
				t.Fatalf("accepted malformed tag: %+v", got)
			}
		})
	}
}

func TestMethodParameterBoundaries(t *testing.T) {
	// The parameter helpers also reject incomplete input when used alone.
	for _, text := range []string{"array<int] $x", "$x = [1", "$x = 'open"} {
		if _, ok := methodParams(text); ok {
			t.Fatalf("accepted incomplete parameters %q", text)
		}
	}
	if _, ok := methodParam("array<int] $x"); ok {
		t.Fatal("accepted mismatched parameter type")
	}
	if got, ok := methodParams(" \t "); !ok || got != nil {
		t.Fatalf("empty parameters = %v, %v", got, ok)
	}
}
