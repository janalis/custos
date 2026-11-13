package types

import (
	"reflect"
	"slices"
	"testing"

	"custos/internal/semantic/wire"
	"custos/internal/semantic/wire/wiretest"
)

// roundTrip encodes t and decodes it again.
func roundTrip(t *testing.T, typ Type) Type {
	t.Helper()
	table := wire.NewStrings()
	w := wire.NewWriter(table)
	typ.Encode(w)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	strs, rest, err := wire.ReadStrings(table.Append(nil))
	if err != nil || len(rest) != 0 {
		t.Fatal(err)
	}
	r := wire.NewReader(w.Buf, strs)
	got := DecodeType(r)
	if !r.Done() {
		t.Fatal("not read whole", r.Err())
	}
	return got
}

// Every field of a type survives encoding, nil slices staying nil and empty
// ones empty.
func TestWireRoundTrip(t *testing.T) {
	for depth := range 5 {
		var typ Type
		wiretest.Fill(&typ, depth, nil)
		if got := roundTrip(t, typ); !reflect.DeepEqual(got, typ) {
			t.Fatalf("depth %d: %#v\n!= %#v", depth, got, typ)
		}
	}
	for _, typ := range []Type{Unknown, Bottom, Int, Ints(IntRange{1, 3}), EnumCases(`\Suit`, "Hearts"), TrackedStrings()} {
		if got := roundTrip(t, typ); !reflect.DeepEqual(got, typ) {
			t.Fatalf("%v: %#v", typ, got)
		}
	}
	empty := Type{atoms: []string{}, gen: []genEntry{}, ints: []IntRange{}, lens: []IntRange{}, arr: &arrayInfo{keys: []ShapeKey{}}, inter: &[]string{}}
	if got := roundTrip(t, empty); !reflect.DeepEqual(got, empty) {
		t.Fatal("empty slices")
	}
}

// The codec writes every field: a field added to a type it reaches fails
// this test until the codec writes it and the list below names it.
func TestWireFields(t *testing.T) {
	want := map[string][]string{
		"types.Type":      {"card", "atoms", "arr", "element", "gen", "inter", "ints", "lens", "cases", "strs"},
		"types.arraySize": {"minimum", "maximum"},
		"types.arrayInfo": {"shape", "sealed", "nonEmpty", "key", "keys", "keyPattern", "elem"},
		"types.ShapeKey":  {"Name", "Type", "Optional"},
		"types.genEntry":  {"atom", "args"},
		"types.IntRange":  {"Lo", "Hi"},
		"types.enumCases": {"class", "names"},
		"types.strFacts":  {"only", "not"},
	}
	got := wiretest.Fields(reflect.TypeFor[Type](), nil, func(reflect.Type) bool { return false })
	if !reflect.DeepEqual(got, want) {
		for name, fields := range got {
			if !slices.Equal(fields, want[name]) {
				t.Errorf("%s: fields %v, the codec writes %v", name, fields, want[name])
			}
		}
		t.Fatal("the codec does not write every field")
	}
}

// A corrupt input decodes without allocating past what it holds.
func TestWireCorrupt(t *testing.T) {
	r := wire.NewReader([]byte{0xff, 0x01}, nil)
	DecodeType(r)
	if r.Err() == nil {
		t.Fatal("corrupt")
	}
	// Nesting past the bound fails both sides.
	deep := Int
	for range wire.MaxDepth {
		inner := deep
		deep = Type{atoms: []string{"array"}, element: &inner}
	}
	w := wire.NewWriter(wire.NewStrings())
	deep.Encode(w)
	if w.Err() == nil {
		t.Fatal("too deep to encode")
	}
	w = wire.NewWriter(wire.NewStrings())
	arr := &arrayInfo{}
	for range wire.MaxDepth {
		arr = &arrayInfo{elem: arr}
	}
	arr.encode(w)
	if w.Err() == nil {
		t.Fatal("array facts too deep to encode")
	}
	// Each level: the element flag, nil atoms, then the next level.
	var b []byte
	for range wire.MaxDepth + 1 {
		b = append(b, wireElement, 0)
	}
	r = wire.NewReader(b, nil)
	DecodeType(r)
	if r.Err() == nil {
		t.Fatal("too deep to decode")
	}
	b = []byte{wireArr, 0}
	for range wire.MaxDepth + 1 {
		b = append(b, 1<<4, 0, 0)
	}
	r = wire.NewReader(b, nil)
	DecodeType(r)
	if r.Err() == nil {
		t.Fatal("array facts too deep to decode")
	}
}
