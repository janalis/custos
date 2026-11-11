package types

import (
	"math"
	"testing"
)

// fakeClasses is a small class hierarchy: \A ← \B (classes), \I ← \C
// (interface, implementing class), \F final, \E enum, \Closure final,
// \Traversable ← \Iter, and \Partial known with an unresolved parent.
type fakeClasses struct{}

var fakeParents = map[string][]string{
	`\A`: nil, `\B`: {`\A`}, `\I`: nil, `\J`: nil, `\C`: {`\I`}, `\F`: nil, `\E`: nil,
	`\Closure`: nil, `\Traversable`: nil, `\Iter`: {`\Traversable`}, `\FinalIter`: {`\Traversable`},
}

func (fakeClasses) Resolved(c string) bool { _, ok := fakeParents[c]; return ok }

func (f fakeClasses) IsSubtype(child, parent string) Verdict {
	if !f.Resolved(child) || !f.Resolved(parent) {
		return Maybe
	}
	for c := child; ; {
		if c == parent {
			return Yes
		}
		ps := fakeParents[c]
		if len(ps) == 0 {
			return No
		}
		c = ps[0]
	}
}

func (fakeClasses) IsFinal(c string) bool {
	return c == `\F` || c == `\E` || c == `\Closure` || c == `\FinalIter`
}

func (fakeClasses) IsInterface(c string) bool { return c == `\I` || c == `\J` || c == `\Traversable` }

func doc(s string) Type { return FromDoc(s, func(w string) string { return w }) }

func TestAccepts(t *testing.T) {
	weak := AcceptContext{Classes: fakeClasses{}}
	strict := AcceptContext{Classes: fakeClasses{}, Strict: true}
	internal := AcceptContext{Classes: fakeClasses{}, Internal: true}
	internalStrict := AcceptContext{Classes: fakeClasses{}, Internal: true, Strict: true}
	noClasses := AcceptContext{Strict: true}
	posInt := Ints(IntRange{1, math.MaxInt64})
	nonEmpty := StringLengths(IntRange{1, math.MaxInt64})
	inter := Of(`\I`, `\J`).withInter([]string{`\I`, `\J`})
	cases := []struct {
		name             string
		expected, actual Type
		ctx              AcceptContext
		want             Verdict
		some             bool // SomeRejected
	}{
		// Unknown, mixed, templates, static/self, unresolved: Maybe.
		{"unknown actual", Int, Unknown, strict, Maybe, false},
		{"unknown expected", Unknown, Int, strict, Maybe, false},
		{"bottom actual", Int, Bottom, strict, Maybe, false},
		{"mixed actual", Int, Mixed, strict, Maybe, false},
		{"mixed expected", Mixed, Of("int", "null"), strict, Yes, false},
		{"nullable mixed expected", Of("mixed", "null"), String, strict, Yes, false},
		{"template actual", Int, Of(`\~T`), strict, Maybe, false},
		{"template expected", Of(`\~~T`), Int, strict, Maybe, false},
		{"static actual", Of(`\A`), Of("static"), strict, Maybe, false},
		{"self expected", Of("self"), Of(`\A`), strict, Maybe, false},
		{"void expected", Void, Int, strict, Maybe, false},
		{"void actual", Int, Void, strict, Maybe, false},
		{"never actual", Int, Of("never"), strict, Yes, false},
		{"unresolved actual class", Of(`\A`), Of(`\Nope`), strict, Maybe, false},
		{"unresolved expected class", Of(`\Nope`), Of(`\A`), strict, Maybe, false},
		{"unresolved expected class, scalar", Of(`\Nope`), Int, strict, Maybe, false},
		{"unresolved actual class, scalar expected", Int, Of(`\Nope`), strict, Maybe, false},
		{"no class relations", Of(`\A`), Of(`\B`), noClasses, Maybe, false},
		{"no class relations, scalar", Of(`\A`), Int, noClasses, Maybe, false},
		{"same class without relations", Of(`\A`), Of(`\A`), noClasses, Yes, false},
		// Scalars, strict mode.
		{"same scalar", Int, Int, strict, Yes, false},
		{"int to float", Float, Int, strict, Yes, false},
		{"int to float weak", Float, Int, weak, Yes, false},
		{"float to int", Int, Float, strict, No, true},
		{"string to int", Int, String, strict, No, true},
		{"int to string", String, Int, strict, No, true},
		{"int to bool", Bool, Int, strict, No, true},
		{"true to bool", Bool, Of("true"), strict, Yes, false},
		{"false to bool", Bool, Of("false"), strict, Yes, false},
		{"bool to true", Of("true"), Bool, strict, Maybe, false},
		{"false to true", Of("true"), Of("false"), strict, No, true},
		{"true to int", Int, Of("true"), strict, No, true},
		{"int or string to int", Int, Of("int", "string"), strict, Maybe, true},
		{"int or float to float", Float, Of("int", "float"), strict, Yes, false},
		// Scalars, weak mode: never No between scalars.
		{"float to int weak", Int, Float, weak, Maybe, false},
		{"string to int weak", Int, String, weak, Maybe, false},
		{"int to string weak", String, Int, weak, Maybe, false},
		{"bool to string weak", String, Bool, weak, Maybe, false},
		{"array to int weak", Int, Array, weak, No, true},
		// null.
		{"null to nullable", Of("int", "null"), Null, strict, Yes, false},
		{"null to int", Int, Null, strict, No, true},
		{"null to int weak", Int, Null, weak, No, true},
		{"nullable to int", Int, Of("int", "null"), strict, Maybe, true},
		{"null to internal int", Int, Null, internal, Maybe, false},
		{"null to internal int strict", Int, Null, internalStrict, No, true},
		{"null to internal array", Array, Null, internal, No, true},
		{"null to class", Of(`\A`), Null, strict, No, true},
		{"null to object", Of("object"), Null, strict, No, true},
		{"null to callable", Of("callable"), Null, strict, No, true},
		// Arrays and iterables.
		{"array to array", Array, Array, strict, Yes, false},
		{"typed array to array", Array, Of("int[]"), strict, Yes, false},
		{"array to iterable", Of("iterable"), Array, strict, Yes, false},
		{"typed array to iterable", Of("iterable"), Of(`\A[]`), strict, Yes, false},
		{"iterable to iterable", Of("iterable"), Of("iterable"), strict, Yes, false},
		{"iterable to array", Array, Of("iterable"), strict, Maybe, false},
		{"callable to array", Array, Of("callable"), strict, Maybe, false},
		{"string to array", Array, String, strict, No, true},
		{"object to array", Array, Of("object"), strict, No, true},
		{"class to array", Array, Of(`\A`), strict, No, true},
		{"array to string weak", String, Array, weak, No, true},
		{"array to class", Of(`\A`), Array, strict, No, true},
		{"empty array to non-empty-array", doc("non-empty-array"), Array.WithShape(nil, true), strict, No, true},
		{"array to non-empty-array", doc("non-empty-array"), Array, strict, Maybe, false},
		{"non-empty to non-empty-array", doc("non-empty-array"), Array.WithNonEmpty(true), strict, Yes, false},
		{"array to shape", doc("array{a: int}"), Array, strict, Maybe, false},
		{"int[] to int[]", doc("int[]"), Of("int[]"), strict, Yes, false},
		{"string[] to int[]", doc("int[]"), Of("string[]"), strict, No, true},
		{"string[] to int[] weak", doc("int[]"), Of("string[]"), weak, No, true},
		{"array to int[]", doc("int[]"), Array, strict, Maybe, false},
		{"empty array to int[]", doc("int[]"), Array.WithShape(nil, true), strict, Yes, false},
		{"B[] to A[]", doc(`A[]`), Of(`\B[]`), strict, Yes, false},
		{"int[] to array<string, int>", doc("array<string, int>"), Of("int[]"), strict, Maybe, false},
		{"int[] to non-empty-list<int>", doc("non-empty-list<int>"), Of("int[]"), strict, Maybe, false},
		{"int to int[]", doc("int[]"), Int, strict, No, true},
		{"iterable to int[]", doc("int[]"), Of("iterable"), strict, Maybe, false},
		{"int[]|string[] to int[]|bool[]", doc("int[]|bool[]"), Of("int[]", "string[]"), strict, Maybe, true},
		{"int[] to iterable<string>", doc("iterable<string>"), Of("int[]"), strict, No, true},
		{"int[] to iterable<int, int>", doc("iterable<int, int>"), Of("int[]"), strict, Yes, false},
		{"Traversable class to iterable", Of("iterable"), Of(`\Iter`), strict, Yes, false},
		{"final non-Traversable to iterable", Of("iterable"), Of(`\F`), strict, No, true},
		{"non-final class to iterable", Of("iterable"), Of(`\A`), strict, Maybe, false},
		{"object to iterable", Of("iterable"), Of("object"), strict, Maybe, false},
		{"callable to iterable", Of("iterable"), Of("callable"), strict, Maybe, false},
		{"string to iterable", Of("iterable"), String, strict, No, true},
		// callable.
		{"Closure to callable", Of("callable"), Of(`\Closure`), strict, Yes, false},
		{"callable to callable", Of("callable"), Of("callable"), strict, Yes, false},
		{"string to callable", Of("callable"), String, strict, Maybe, false},
		{"array to callable", Of("callable"), Array, strict, Maybe, false},
		{"object to callable", Of("callable"), Of("object"), strict, Maybe, false},
		{"class to callable", Of("callable"), Of(`\A`), strict, Maybe, false},
		{"int to callable", Of("callable"), Int, strict, No, true},
		{"callable to Closure", Of(`\Closure`), Of("callable"), strict, Maybe, false},
		{"callable to string", String, Of("callable"), strict, Maybe, false},
		{"callable to int", Int, Of("callable"), strict, No, true},
		// object.
		{"class to object", Of("object"), Of(`\A`), strict, Yes, false},
		{"object to object", Of("object"), Of("object"), strict, Yes, false},
		{"callable to object", Of("object"), Of("callable"), strict, Maybe, false},
		{"iterable to object", Of("object"), Of("iterable"), strict, Maybe, false},
		{"int to object", Of("object"), Int, strict, No, true},
		{"object to class", Of(`\A`), Of("object"), strict, Maybe, false},
		{"iterable to class", Of(`\Traversable`), Of("iterable"), strict, Maybe, false},
		{"object to string weak", String, Of("object"), weak, Maybe, false},
		{"object to string strict", String, Of("object"), strict, No, true},
		{"class to string weak", String, Of(`\A`), weak, Maybe, false},
		{"class to string strict", String, Of(`\A`), strict, No, true},
		{"class to int weak", Int, Of(`\A`), weak, No, true},
		// Resources: stubs may lag behind objects.
		{"resource to resource", Of("resource"), Of("resource"), strict, Yes, false},
		{"class to resource", Of("resource"), Of(`\A`), strict, Maybe, false},
		{"resource to class", Of(`\A`), Of("resource"), strict, Maybe, false},
		// Classes.
		{"same class", Of(`\A`), Of(`\A`), strict, Yes, false},
		{"subclass", Of(`\A`), Of(`\B`), strict, Yes, false},
		{"implementer", Of(`\I`), Of(`\C`), strict, Yes, false},
		{"superclass to class", Of(`\B`), Of(`\A`), strict, No, true},
		{"unrelated classes", Of(`\A`), Of(`\C`), strict, No, true},
		{"final class to interface", Of(`\I`), Of(`\F`), strict, No, true},
		{"enum to interface", Of(`\I`), Of(`\E`), strict, No, true},
		{"non-final class to interface", Of(`\I`), Of(`\A`), strict, Maybe, false},
		{"interface to class", Of(`\A`), Of(`\I`), strict, Maybe, false},
		{"interface to interface", Of(`\J`), Of(`\I`), strict, Maybe, false},
		{"class or null to class", Of(`\A`), Of(`\B`, "null"), strict, Maybe, true},
		{"class to int|class", Of("int", `\A`), Of(`\B`), strict, Yes, false},
		{"int to class", Of(`\A`), Int, strict, No, true},
		{"string to class weak", Of(`\A`), String, weak, No, true},
		{"enum cases", EnumCases(`\E`, "X", "Y"), EnumCases(`\E`, "X"), strict, Yes, false},
		{"enum other case", EnumCases(`\E`, "X"), EnumCases(`\E`, "Y"), strict, No, true},
		{"enum without cases", EnumCases(`\E`, "X"), Of(`\E`), strict, Maybe, false},
		// Intersections.
		{"intersection to its interface", Of(`\I`), inter, strict, Yes, false},
		{"intersection to object", Of("object"), inter, strict, Yes, false},
		{"intersection to unrelated class", Of(`\A`), inter, strict, Maybe, false},
		{"intersection or null to interface", Of(`\I`), Union(inter, Null), strict, Maybe, true},
		{"intersection or null to nullable interface", Of(`\I`, "null"), Union(inter, Null), strict, Yes, false},
		{"int to null", Null, Int, strict, No, true},
		{"'a'|'c' to non-empty 'a'|'b'", stringValues("a", "b").withLens([]IntRange{{1, math.MaxInt64}}), stringValues("a", "c"), strict, Maybe, false},
		{"class to intersection", inter, Of(`\C`), strict, Maybe, false},
		{"final class to intersection", inter, Of(`\F`), strict, No, true},
		{"int to intersection", inter, Int, strict, No, true},
		{"object to intersection", inter, Of("object"), strict, Maybe, false},
		{"object to nullable intersection", Union(inter, Null), Of("object"), strict, Maybe, false},
		{"null to nullable intersection", Union(inter, Null), Null, strict, Yes, false},
		// Refinements count only against exact values.
		{"int to positive-int", posInt, Int, strict, Maybe, false},
		{"5 to positive-int", posInt, Ints(IntRange{5, 5}), strict, Yes, false},
		{"0 to positive-int", posInt, Ints(IntRange{0, 0}), strict, No, true},
		{"-1..1 to positive-int", posInt, Ints(IntRange{-1, 1}), strict, Maybe, false},
		{"5 to int", Int, Ints(IntRange{5, 5}), strict, Yes, false},
		{"string to non-empty-string", nonEmpty, String, strict, Maybe, false},
		{"length 3 to non-empty-string", nonEmpty, StringLengths(IntRange{3, 3}), strict, Yes, false},
		{"empty string to non-empty-string", nonEmpty, StringLengths(IntRange{0, 0}), strict, No, true},
		{"'' to non-empty-string", nonEmpty, stringValues(""), strict, No, true},
		{"'ab' to non-empty-string", nonEmpty, stringValues("ab"), strict, Yes, false},
		{"'a' to 'a'|'b'", stringValues("a", "b"), stringValues("a"), strict, Yes, false},
		{"'c' to 'a'|'b'", stringValues("a", "b"), stringValues("c"), strict, No, true},
		{"'a'|'c' to 'a'|'b'", stringValues("a", "b"), stringValues("a", "c"), strict, Maybe, false},
		{"string to 'a'|'b'", stringValues("a", "b"), String, strict, Maybe, false},
		{"'' to non-empty 'a'|''", stringValues("a", "").withLens([]IntRange{{1, math.MaxInt64}}), stringValues(""), strict, No, true},
		{"5 to float", Float, Ints(IntRange{5, 5}), strict, Yes, false},
		{"int to int|non-empty-string", Union(Int, nonEmpty), Int, strict, Yes, false},
		{"float to float|positive-int", Union(Float, posInt), Float, strict, Yes, false},
	}
	for _, c := range cases {
		if got := Accepts(c.expected, c.actual, c.ctx); got != c.want {
			t.Errorf("%s: Accepts(%s, %s) = %s, want %s", c.name, c.expected.ShapeString(), c.actual.ShapeString(), got, c.want)
		}
		if got := AllRejected(c.expected, c.actual, c.ctx); got != (c.want == No) {
			t.Errorf("%s: AllRejected = %v", c.name, got)
		}
		if got := SomeRejected(c.expected, c.actual, c.ctx); got != c.some {
			t.Errorf("%s: SomeRejected = %v, want %v", c.name, got, c.some)
		}
	}
}

func TestVerdictString(t *testing.T) {
	for v, want := range map[Verdict]string{Yes: "yes", No: "no", Maybe: "maybe"} {
		if v.String() != want {
			t.Errorf("%d: %s", v, v)
		}
	}
}

func BenchmarkAccepts(b *testing.B) {
	ctx := AcceptContext{Classes: fakeClasses{}, Strict: true}
	cases := []struct {
		name             string
		expected, actual Type
	}{
		{"scalar", Int, Int},
		{"nullable", Int, Of("int", "null")},
		{"class", Of(`\A`), Of(`\B`)},
		{"union", Of("int", "string", `\A`, "null"), Of(`\C`, "float")},
		{"array", doc("int[]"), Of("string[]")},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				Accepts(c.expected, c.actual, ctx)
			}
		})
	}
}
