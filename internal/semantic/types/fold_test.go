package types

import "testing"

// normalizeAtom folds case without allocating; non-ASCII names go through
// strings.ToLower, so its Unicode folding is kept.
func TestNormalizeAtomFold(t *testing.T) {
	for in, want := range map[string]string{
		`\Int`:                    "int",
		`\Foo`:                    `\Foo`,
		"INTEGER":                 "int",
		"String[]":                "string[]",
		"Foo[]":                   "Foo[]",
		"callbacK":                "callable", // Kelvin sign folds to k
		"Ünknown":                 "Ünknown",
		`\AVeryLongClassNameHere`: `\AVeryLongClassNameHere`,
	} {
		if got := normalizeAtom(in); got != want {
			t.Errorf("normalizeAtom(%q) = %q, want %q", in, got, want)
		}
	}
}

func BenchmarkNormalizeAtomClass(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		normalizeAtom(`\Stringable`)
		normalizeAtom("Integer")
	}
}
