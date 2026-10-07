package phpdoc

import "testing"

func TestSplitType(t *testing.T) {
	cases := []struct{ in, typ, rest string }{
		{"int | null the id", "int|null", "the id"},
		{"array<string, int> $map", "array<string,int>", "$map"},
		{"(TKey is null ? array<string, mixed> : array<mixed>) all values", "(TKey is null ? array<string, mixed> : array<mixed>)", "all values"},
	}
	for _, c := range cases {
		typ, rest := SplitType(c.in)
		if typ != c.typ || rest != c.rest {
			t.Errorf("SplitType(%q) = %q, %q; want %q, %q", c.in, typ, rest, c.typ, c.rest)
		}
	}
}
