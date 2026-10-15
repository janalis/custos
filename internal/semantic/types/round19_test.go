package types

import "testing"

func TestRound19DocTypes(t *testing.T) {
	for expr, want := range map[string]string{
		"Status::Active|null": "mixed|null", "Foo::*": "mixed", "array{a:int}['a']": "int", "array{a:int}['b']": "mixed", "T[K]": "mixed",
		"key-of<array{a:int,2:string}>": "int|string", "value-of<array{a:int,2:string}>": "int|string", "value-of<array<int,string>>": "string", "key-of<Foo>": "mixed",
		"key-of": "mixed", "value-of<": "mixed", "non-empty-literal-string": "string", "non-empty-lowercase-string": "string", "uppercase-string": "string", "callable-object": "callable|object", "pure-callable": "callable", "pure-Closure": `\Closure`, "non-empty-mixed": "mixed",
	} {
		if got := FromDoc(expr, nil).String(); got != want {
			t.Errorf("%s = %s, want %s", expr, got, want)
		}
	}
	resolve := func(expr string) Type {
		if expr == "Status::Active" {
			return Of(`\Status`)
		}
		if expr == "value-of<Status>" {
			return String
		}
		return Unknown
	}
	for expr, want := range map[string]string{"Status::Active|null": `\Status|null`, "Foo::*": "mixed", "value-of<Status>": "string", "key-of<Foo>": "mixed"} {
		if got := FromDocResolved(expr, nil, resolve).String(); got != want {
			t.Errorf("resolved %s = %s want %s", expr, got, want)
		}
	}
}
