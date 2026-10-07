package types

import (
	"strings"
	"testing"
)

func TestCallableSignatures(t *testing.T) {
	cases := map[string]string{
		"callable(int): string":             "string",
		"Closure(int, string): ?Foo":        `\Foo|null`,
		`\Closure(): (int|false)`:           "false|int",
		"callable(): callable(): int":       "callable",
		"?Closure(): array{a: int}":         "array",
		"callable":                          "?unknown",
		"callable(int)":                     "?unknown", // no return type
		"callable(int":                      "?unknown", // unbalanced
		"callable(): mixed":                 "?unknown",
		"callable(): T":                     "?unknown", // a template reads as mixed
		"callable(): int|Closure(): string": "?unknown", // members disagree
		"callable(): int|Closure(): int":    "int",
		"Closure":                           "?unknown",
	}
	resolve := func(w string) string {
		if w == "T" {
			return ""
		}
		return strings.TrimPrefix(w, `\`)
	}
	for in, want := range cases {
		got := FromDoc(in, resolve)
		if r := CallableReturn(got); r.String() != want {
			t.Errorf("%s: return %s, want %s", in, r, want)
		}
		if back := FromDoc(got.DocString(), nil); CallableReturn(back).DocString() != CallableReturn(got).DocString() {
			t.Errorf("%s: DocString %q does not round-trip", in, got.DocString())
		}
	}
	if r := CallableReturn(FromDoc("callable(): callable(): int", nil)); CallableReturn(r).String() != "int" {
		t.Errorf("nested signature: %s", r.DocString())
	}
	c := WithCallableReturn(Of(`\Closure`), `\Closure`, Mixed)
	if c.TypeArgs(`\Closure`) != nil {
		t.Error("a mixed return is not recorded")
	}
	if got := FromDoc("(T is ?int ? callable(): int : string)", nil).String(); got != "callable|string" {
		t.Errorf("conditional with `?int` target and a callable branch: %s", got)
	}
	if got := FromDoc("(T is int?A:B)", nil).String(); got != `\A|\B` {
		t.Errorf("unspaced conditional: %s", got)
	}
}

func TestParseCond(t *testing.T) {
	resolve := func(w string) string {
		switch w {
		case "T":
			return ""
		case "Alias":
			return "=int|string"
		}
		return "App\\" + strings.TrimPrefix(w, `\`)
	}
	subject := func(w string) string {
		if w == "T" {
			return "v"
		}
		return ""
	}
	cases := map[string]string{
		"($x is string ? int : float)":                     "($x is string ? int : float)",
		"($x is not null ? Foo : null)":                    `($x is not null ? \App\Foo : null)`,
		"( $x is 'a' ? int : ($x is 2 ? float : string) )": "($x is 'a' ? int : ($x is 2 ? float : string))",
		"($m is Foo::BAR ? array<string> : false)":         `($m is \App\Foo::BAR ? string[] : false)`,
		"($x is non-empty-string ? int : float)":           "($x is ~string ? int : float)",
		"(T is array ? list<string> : string)":             "($v is array ? string[] : string)",
		"($x is scalar|array-key|?Foo ? int : float)":      `($x is \App\Foo|bool|float|int|null|string ? int : float)`,
		"($x is T ? int : float)":                          "($x is ~mixed ? int : float)",
		"($x is Alias ? int : float)":                      "($x is ~int|string ? int : float)",
		"($x is -1.5 ? int : float)":                       "($x is -1.5 ? int : float)",
		"($x is integer ? int : float)":                    "($x is int ? int : float)",
		"($x is static ? int : float)":                     "($x is ~static ? int : float)",
	}
	for in, want := range cases {
		c, ok := ParseCond(in, resolve, subject)
		if !ok {
			t.Errorf("%s: does not parse", in)
			continue
		}
		if got := c.String(); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
		back, ok := ParseCond(c.String(), nil, nil)
		if !ok || back.String() != c.String() || back.Exact != c.Exact {
			t.Errorf("%s: canonical form %s does not round-trip", in, c.String())
		}
	}
	bad := []string{
		"int",                           // not a conditional
		"($x is string ? int)",          // no else branch
		"(is string ? int : float)",     // no subject
		"(U is string ? int : float)",   // a subject that is no parameter
		"($1x is string ? int : float)", // not an identifier
		"($x is 'a\\'b' ? int : float)", // escapes in a string target
		"($x is 'ab ? int : float)",     // unterminated
		`($x is 'a"b' ? int : float)`,   // a quote in a string target
		"($x-y is int ? int : float)",   // not an identifier
		"($x is 1.2.3 ? int : float)",   // not a number
		"($x is Foo::1 ? int : float)",  // not a constant name
		"($x is T::A ? int : float)",    // a template's constant
		"($x is {x} ? int : float)",     // not a type
		"($x is int ? {x} : float)",     // a branch that is not a type
		"($x is int ? int : ($x is int ? {x} : float))",
		"($x is int ? int : float) | null", // not wrapped as a whole
		strings.Repeat("(", 40) + "$x is int ? int : float" + strings.Repeat(")", 40),
		"(" + strings.Repeat("a", MaxDocTypeLen) + ")",
	}
	for _, in := range bad {
		if c, ok := ParseCond(in, resolve, subject); ok {
			t.Errorf("%q: parsed as %s", in, c)
		}
	}
	// Nesting beyond the depth cap.
	deep := "int"
	for i := 0; i < 40; i++ {
		deep = "($x is string ? float : " + deep + ")"
	}
	if _, ok := ParseCond(deep, nil, nil); ok {
		t.Error("nesting beyond the cap parsed")
	}
	if !isNumberLiteral("-12.5") || isNumberLiteral("-") || isNumberLiteral("1.2.3") {
		t.Error("isNumberLiteral")
	}
}
