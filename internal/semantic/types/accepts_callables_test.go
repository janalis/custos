package types

import "testing"

func TestAcceptsCallableSignatures(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	for _, c := range []struct {
		expected, actual string
		want             Verdict
	}{
		// The return type is covariant.
		{`callable(): \A`, `callable(): \B`, Yes},
		{`callable(): \B`, `callable(): \F`, No},
		{`callable(): array`, `callable(): int`, No},
		// Parameters are contravariant.
		{`callable(\B): void`, `callable(\A): void`, Yes},
		{`callable(\F): void`, `callable(\B): void`, No},
		{`callable(array): void`, `callable(int): void`, No},
		// Scalars may be coerced where the callable is declared.
		{`callable(): string`, `callable(): int`, Yes},
		{`callable(string): void`, `callable(int): void`, Yes},
		// Arity: the callable may not require more than is passed.
		{`callable(int): void`, `callable(int, int): void`, No},
		{`callable(): void`, `callable(int): void`, No},
		{`callable(int): void`, `callable(int, int=): void`, Yes},
		{`callable(int): void`, `callable(int, int ...$rest): void`, Yes},
		{`callable(int ...$all): void`, `callable(int, int): void`, Yes},
		{`callable(int, int): void`, `callable(int): void`, Yes},
		// Closures take part on either side.
		{`Closure(): array`, `Closure(): int`, No},
		{`callable(): array`, `Closure(): int`, No},
		{`Closure(int): void`, `Closure(int, int): void`, No},
		// Nothing written decides nothing.
		{`callable`, `callable(): int`, Yes},
		{`callable(): array`, `callable`, Yes},
		{`callable(mixed): mixed`, `callable(int): int`, Yes},
	} {
		if got := Accepts(FromDoc(c.expected, nil), FromDoc(c.actual, nil), ctx); got != c.want {
			t.Errorf("Accepts(%s, %s) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

// An inferred closure knows its return type but not its parameters.
func TestAcceptsCallableWithoutShape(t *testing.T) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	inferred := WithCallableReturn(Of(`\Closure`), `\Closure`, Int)
	if got := Accepts(FromDoc(`callable(int, int): int`, nil), inferred, ctx); got != Yes {
		t.Errorf("arity unknown = %v, want yes", got)
	}
	if got := Accepts(FromDoc(`callable(): array`, nil), inferred, ctx); got != No {
		t.Errorf("return type = %v, want no", got)
	}
}

func TestAcceptsNestedCallablesAreBounded(t *testing.T) {
	want, have := "array", "int"
	for range maxDocDepth - 2 {
		want, have = "callable(): "+want, "callable(): "+have
	}
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	if got := Accepts(FromDoc(want, nil), FromDoc(have, nil), ctx); got == No {
		t.Errorf("a mismatch beyond the depth limit decided: %v", got)
	}
}
