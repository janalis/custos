package types

import "testing"

func TestAcceptsGenericArguments(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	for _, c := range []struct {
		expected, actual string
		want             Verdict
	}{
		// Invariant: disjoint arguments never fit, either way round.
		{`\A<int>`, `\A<string>`, No},
		{`\A<int>`, `\A<int>`, Yes},
		{`\A<int|string>`, `\A<int>`, Yes},
		{`\A<int>`, `\A<int|string>`, Yes},
		// Covariant: only the accepting direction counts.
		{`\B<int>`, `\B<string>`, No},
		{`\B<\A>`, `\B<\B>`, Yes},
		{`\B<\B>`, `\B<\A>`, Yes},
		// Contravariant: the argument accepts the expected one.
		{`\C<int>`, `\C<string>`, No},
		{`\C<\B>`, `\C<\A>`, Yes},
		// Every position is compared.
		{`\F<int, string>`, `\F<int, int>`, No},
		{`\F<int, string>`, `\F<string, string>`, No},
		// Nothing to compare.
		{`\A`, `\A<string>`, Yes},
		{`\A<int>`, `\A`, Yes},
		{`\F<int, string>`, `\F<int>`, Yes},
		{`\A<int, string>`, `\A<int, int>`, Yes}, // no template at position 1
		{`\J<int>`, `\J<string>`, Yes},           // declares no template
		{`\A<mixed>`, `\A<string>`, Yes},
		{`\A<\~T>`, `\A<string>`, Yes},
		{`\A<int>`, `\A<\~T>`, Yes},
		{`\A<\Missing>`, `\A<string>`, Yes},
		// A subclass is not mapped onto its parent's templates.
		{`\A<int>`, `\B<string>`, Yes},
	} {
		if got := Accepts(FromDoc(c.expected, nil), FromDoc(c.actual, nil), ctx); got != c.want {
			t.Errorf("Accepts(%s, %s) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

func TestAcceptsGenericArgumentsWithoutClasses(t *testing.T) {
	if got := Accepts(FromDoc(`\A<int>`, nil), FromDoc(`\A<string>`, nil), AcceptContext{Strict: true}); got != Yes {
		t.Errorf("no class relations: %v", got)
	}
}

func TestAcceptsGenericArgumentInUnion(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	want, have := FromDoc(`\A<int>|null`, nil), FromDoc(`\A<string>|null`, nil)
	if got := Accepts(want, have, ctx); got != Maybe {
		t.Errorf("nullable = %v, want maybe", got)
	}
	if !SomeRejected(want, have, ctx) {
		t.Error("the mismatching member is not reported")
	}
}

func TestWithoutClassTypeArgs(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{`\A<int>`, `\A`},
		{`\A<int>|null`, `\A|null`},
		{`array<int, \A<int>>`, `array<int,\A>`},
		{`callable(int): string`, `callable(int): (string)`},
		{`Closure(int): string`, `\Closure(int): (string)`},
		{`iterable<int, string>`, `iterable<int, string>`},
		{`class-string<\A>`, `class-string<\A>`},
		{`int`, `int`},
	} {
		if got := FromDoc(c.in, nil).WithoutClassTypeArgs().DocString(); got != c.want {
			t.Errorf("%s = %s, want %s", c.in, got, c.want)
		}
	}
}
