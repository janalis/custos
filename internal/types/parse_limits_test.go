package types

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"custos/internal/testbudget"
)

// Doc types come from analysed code: pathological input must parse fast.
func TestFromDocPathological(t *testing.T) {
	aliases := map[string]string{"A": "B|B", "B": "array{x: A, y: A}|A[]"}
	// Each alias uses the next one three times: naive expansion is 3^n.
	for i := 0; i < 60; i++ {
		aliases["E"+strconv.Itoa(i)] = strings.Repeat("E"+strconv.Itoa(i+1)+"|", 2) + "array<E" + strconv.Itoa(i+1) + ">"
	}
	// A huge alias with few parts, used many times: each expansion used to
	// rescan the whole definition (2,000 x 1 MB).
	aliases["Huge"] = "array{k" + strings.Repeat("x", 1<<20) + ": int}"
	aliases["ManyHuge"] = strings.TrimSuffix(strings.Repeat("Huge|", 2000), "|")
	// An alias whose members are long but under the cap, used many times:
	// bounded by the byte budget.
	aliases["Long"] = "array{k" + strings.Repeat("x", MaxDocTypeLen-20) + ": int}"
	aliases["ManyLong"] = strings.TrimSuffix(strings.Repeat("Long|", 800), "|")
	defs := map[string]string{}
	for k, d := range aliases {
		defs[k] = "=" + d // built once: the test resolver must not dominate
	}
	resolve := func(w string) string {
		if d, ok := defs[w]; ok {
			return d
		}
		return w
	}
	nest := func(open, mid, close string, n int) string {
		return strings.Repeat(open, n) + mid + strings.Repeat(close, n)
	}
	cases := map[string]string{
		"generics 10k":      nest("array<", "int", ">", 10000),
		"generics in cap":   nest("array<", "int", ">", 580),
		"lists in cap":      nest("list<", "int", ">", 680),
		"brackets 10k":      "int" + strings.Repeat("[]", 10000),
		"brackets in cap":   "int" + strings.Repeat("[]", 2000),
		"shapes 10k":        nest("array{a: ", "int", "}", 10000),
		"shapes in cap":     nest("array{a: ", "int", "}", 400),
		"parens in cap":     nest("(", "int", ")", 2000),
		"nullable in cap":   strings.Repeat("?", 4000) + "int",
		"generic args":      "array<" + strings.Repeat("array<int, string>|", 200) + "int>",
		"iterables in cap":  nest("iterable<", "int", ">", 400),
		"conditionals":      nest("(T is int ? ", "int", " : false)", 150),
		"mutual aliases":    "A",
		"exploding aliases": "E0",
		"huge alias":        "ManyHuge",
		"long aliases":      "ManyLong",
	}
	for name, in := range cases {
		start := time.Now()
		got := FromDoc(in, resolve)
		_ = got.DocString()
		if d := time.Since(start); d > testbudget.Of(200*time.Millisecond) && !raceEnabled {
			t.Errorf("%s: took %v", name, d)
		}
	}
	if !FromDoc(strings.Repeat("a", MaxDocTypeLen+1), nil).IsUnknown() {
		t.Error("over-long doc type must be unknown")
	}
	// Deep but legal nesting still parses up to the cap.
	if got := FromDoc(nest("array<", "int", ">", 3), nil).String(); got != "int[][][]" {
		t.Errorf("array<array<array<int>>>: got %s", got)
	}
}

// FuzzFromDoc checks that doc type parsing never panics, stays fast and
// that DocString round-trips to the same atoms.
func FuzzFromDoc(f *testing.F) {
	for _, s := range []string{
		"int|null", "array{a: int, b?: list<string>}", "?Foo[]", "(T is int ? A : B)",
		"array<int, array{x: Alias}>", "non-empty-list<int>", "Foo&Bar", "callable(int): void", "iterable<int, Foo>",
		"Collection<int, Foo>|Foo[]", "array{...}", "'a'|1|1.5",
		"($x is 'a' ? int : ($x is Foo::BAR ? ?int : callable(): string))", "(T is not non-empty-string ? A : B)",
		strings.Repeat("Alias|", 600) + "Alias", "array{a: Alias, b: list<Alias>, c: Alias2}",
	} {
		f.Add(s)
	}
	resolve := func(w string) string {
		switch w {
		case "Alias":
			return "=array{a: Alias, b: Alias[]}|Alias2"
		case "Alias2":
			return "=list<Alias>"
		case "T":
			return ""
		}
		return strings.TrimPrefix(w, `\`)
	}
	f.Fuzz(func(t *testing.T, s string) {
		start := time.Now()
		got := FromDoc(s, resolve)
		c, condOK := ParseCond(s, resolve, func(string) string { return "x" })
		if d := time.Since(start); d > testbudget.Of(time.Second) && !raceEnabled {
			t.Fatalf("%q took %v", s, d)
		}
		if condOK {
			// The canonical form parses back to itself.
			if cs := c.String(); len(cs) <= MaxDocTypeLen {
				if back, ok := ParseCond(cs, nil, nil); !ok || back.String() != cs {
					t.Fatalf("%q: canonical %q does not parse back", s, cs)
				}
			}
		}
		ds := got.DocString()
		if got.IsUnknown() || len(ds) > MaxDocTypeLen || strings.ContainsAny(s, "'\"") {
			return
		}
		if !wellFormedType(got) {
			return // garbage names need not round-trip
		}
		if back := FromDoc(ds, nil); !back.Equal(got) {
			t.Fatalf("%q: DocString %q parses back as %s, want %s", s, ds, back, got)
		}
	})
}

// wellFormedType reports whether every atom of t, its shape keys' and its
// generic arguments' types is well formed.
func wellFormedType(t Type) bool {
	for _, a := range t.Atoms() {
		if !wellFormedAtom(a) {
			return false
		}
	}
	for _, g := range t.gen {
		for _, x := range g.args {
			if !x.IsUnknown() && !wellFormedType(x) {
				return false
			}
		}
	}
	for _, k := range t.ShapeKeys() {
		if !k.Type.IsUnknown() && !wellFormedType(k.Type) {
			return false
		}
	}
	return true
}

// wellFormedAtom reports an atom made of a (backslash-separated) identifier
// and `[]` suffixes.
func wellFormedAtom(a string) bool {
	a = strings.TrimLeft(a, `\`)
	for strings.HasSuffix(a, "[]") {
		a = a[:len(a)-2]
	}
	if a == "" {
		return false
	}
	for i := 0; i < len(a); i++ {
		c := a[i]
		ident := c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
		if !ident && !(i > 0 && (c >= '0' && c <= '9' || c == '\\' || c == '-')) {
			return false
		}
	}
	return !strings.HasSuffix(a, `\`) && !strings.Contains(a, `\\`)
}
