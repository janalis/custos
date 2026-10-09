package types

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestFromDocEdges(t *testing.T) {
	cases := map[string]string{
		"{a}|$x":                   "?",   // no member is a type
		"int|":                     "int", // empty member
		"?":                        "?",   // nullable of an unknown type
		"{a}[]":                    "?",   // element of an unknown type
		"(A ? B : C)":              "?",   // not a conditional type
		"(T is int ? A)":           "?",
		"(T is int ? : B)":         "?",
		`Foo\1x`:                   "?",     // segment starting with a digit
		"array{a: int, a: string}": "array", // duplicate key: no shape
	}
	for in, want := range cases {
		got := FromDoc(in, nil)
		s := got.String()
		if got.IsUnknown() {
			s = "?"
		}
		if s != want {
			t.Errorf("%q: got %s want %s", in, s, want)
		}
	}
	if FromDoc("array{a: int, a: string}", nil).HasShape() {
		t.Error("duplicate keys must drop the shape")
	}
	// A key that is not identifier-like reads as a positional entry.
	if got := FromDoc("array{a b: int}", nil).ShapeString(); got != "array{0: mixed}" && !strings.HasPrefix(got, "array{0:") {
		t.Errorf("non-key prefix: %s", got)
	}
	if got := FromDoc("array{?: int}", nil).ShapeString(); !strings.HasPrefix(got, "array{0:") {
		t.Errorf("empty key: %s", got)
	}
	// More keys than MaxShapeKeys: no shape, but required keys still make it non-empty.
	var b strings.Builder
	b.WriteString("array{")
	for i := 0; i <= MaxShapeKeys; i++ {
		fmt.Fprintf(&b, "k%d: int,", i)
	}
	b.WriteString("}")
	big := FromDoc(b.String(), nil)
	if big.HasShape() || !big.IsNonEmptyArray() {
		t.Errorf("big shape: %s", big.ShapeString())
	}
	opt := strings.ReplaceAll(b.String(), ":", "?:")
	if FromDoc(opt, nil).IsNonEmptyArray() {
		t.Error("big shape of optional keys must not be non-empty")
	}
}

func TestFromNodeOther(t *testing.T) {
	if got := FromNode(&syntax.Variable{}, nil); !got.IsUnknown() {
		t.Errorf("non-type node: %v", got.Atoms())
	}
}

func TestOfEmpty(t *testing.T) {
	if !Of().IsUnknown() {
		t.Error("Of() is unknown")
	}
	if Int.Without("int").String() != "?unknown" {
		t.Error("an emptied type prints as unknown")
	}
	if !Of(" ", "").IsUnknown() {
		t.Error("blank atoms are dropped")
	}
}

func TestGenericEdges(t *testing.T) {
	c := FromDoc("Collection<int>|Box<string>", nil)
	if a := c.TypeArgs(`\Box`); len(a) != 1 || a[0].String() != "string" {
		t.Fatalf("Box args: %v", a)
	}
	// Replacing one atom's arguments keeps the other's.
	r := c.WithTypeArgs(`\Collection`, []Type{Float})
	if r.TypeArgs(`\Box`) == nil || r.TypeArgs(`\Collection`)[0].String() != "float" {
		t.Errorf("replace: %s", r.ShapeString())
	}
	if c.WithTypeArgs(`\Missing`, []Type{Int}).ShapeString() != c.ShapeString() {
		t.Error("absent atom: unchanged")
	}
	// withGen drops entries for absent atoms or without arguments, and keeps the first of duplicates.
	w := Of(`\A`).withGen([]genEntry{{atom: `\B`, args: []Type{Int}}, {atom: `\A`}, {atom: `\A`, args: []Type{Int}}, {atom: `\A`, args: []Type{String}}})
	if a := w.TypeArgs(`\A`); len(a) != 1 || a[0].String() != "int" {
		t.Errorf("withGen: %s", w.ShapeString())
	}
	// A conflicting member drops only its atom.
	m := FromDoc("A<int>|A<string>|B<int>", nil)
	if m.TypeArgs(`\A`) != nil || m.TypeArgs(`\B`) == nil {
		t.Errorf("mergeGen: %s", m.ShapeString())
	}
	if sameArgs([]Type{Int}, []Type{Int, Int}) {
		t.Error("sameArgs length")
	}
	if !ClassString(Int).Equal(String) || ClassString(Int).TypeArgs("string") != nil {
		t.Error("ClassString of a non-class is a plain string")
	}
	if !ClassStringOf(Int).IsUnknown() || !ClassStringOf(Of("string", "int")).IsUnknown() {
		t.Error("ClassStringOf of a non-string")
	}
}

func TestShapeEdges(t *testing.T) {
	if _, ok := Array.ShapeKey("a"); ok {
		t.Error("ShapeKey without shape")
	}
	s := FromDoc("array{a: int}", nil)
	if _, ok := s.ShapeKey("b"); ok {
		t.Error("absent key")
	}
	if s.WithoutArrayInfo().HasShape() || !s.WithoutArrayInfo().Equal(Array) {
		t.Error("WithoutArrayInfo")
	}
	if Array.MapShape(func(k ShapeKey) ShapeKey { return k }).HasShape() {
		t.Error("MapShape without shape")
	}
	// WithElem attaches element facts.
	el := Of("array[]").WithElem(s)
	if got := el.ShapeString(); got != "array[]<array{a: int}>" {
		t.Errorf("WithElem: %s", got)
	}
	if Of("int[]").WithElem(Int).ShapeString() != "int[]" {
		t.Error("WithElem without facts: unchanged")
	}
	// mergeInfo of the same facts.
	if a := mergeInfo(s.arr, s.arr); a != s.arr {
		t.Error("mergeInfo identity")
	}
	// Merging two sealed shapes whose keys together exceed the cap drops the shape.
	var a, b strings.Builder
	a.WriteString("array{")
	b.WriteString("array{")
	for i := 0; i < MaxShapeKeys; i++ {
		fmt.Fprintf(&a, "a%d: int,", i)
		fmt.Fprintf(&b, "b%d: int,", i)
	}
	a.WriteString("}")
	b.WriteString("}")
	u := Union(FromDoc(a.String(), nil), FromDoc(b.String(), nil))
	if u.HasShape() || !u.IsNonEmptyArray() {
		t.Errorf("merged over cap: %s", u.ShapeString())
	}
	// A member without facts drops the union's facts.
	if unionInfo([]Type{s, Of("array", "int[]").WithElem(s), Array}) != nil {
		t.Error("unionInfo with a bare member")
	}
	if unionInfo([]Type{Of("array[]").WithElem(s), Of("int[]").WithNonEmpty(true)}) != nil {
		t.Error("unionInfo: incompatible facts")
	}
	// Non-empty shape whose keys are all optional.
	ne := FromDoc("non-empty-array{a?: int}", nil)
	if got := ne.ShapeString(); got != "non-empty array{a?: int}" {
		t.Errorf("ShapeString: %s", got)
	}
	if got := ne.DocString(); got != "non-empty-array{a?:int}" {
		t.Errorf("DocString: %s", got)
	}
	// `array` member with only element facts.
	if got := Of("array", "array[]").WithElem(s).DocString(); got != "array|array<array{a:int}>" {
		t.Errorf("DocString elem-only: %s", got)
	}
	// Key quoting.
	if got := FromDoc(`array{"it's": int}`, nil).DocString(); got != `array{"it's":int}` {
		t.Errorf("quoted key: %s", got)
	}
}
