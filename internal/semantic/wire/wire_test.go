package wire

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

// encode writes values with w and returns the whole encoding: the string
// table, then the body.
func encode(write func(w *Writer)) []byte {
	table := NewStrings()
	w := NewWriter(table)
	write(w)
	return append(table.Append(nil), w.Buf...)
}

// decode returns a reader of an encoding encode returned.
func decode(t *testing.T, b []byte) *Reader {
	t.Helper()
	table, rest, err := ReadStrings(b)
	if err != nil {
		t.Fatal(err)
	}
	return NewReader(rest, table)
}

func TestRoundTrip(t *testing.T) {
	b := encode(func(w *Writer) {
		w.Uint(math.MaxUint64)
		w.Int(-7)
		w.Float(math.Inf(-1))
		w.Bool(true)
		w.Bool(false)
		w.Bools(true, false, true)
		w.String("a")
		w.String("b")
		w.String("a")
		w.Strings(nil)
		w.Strings([]string{})
		w.Strings([]string{"b", "c"})
		w.Len(2, false)
		w.Uint(0)
		w.Uint(0)
	})
	r := decode(t, b)
	var x, y, z bool
	if r.Uint() != math.MaxUint64 || r.Int() != -7 || r.Float() != math.Inf(-1) || !r.Bool() || r.Bool() {
		t.Fatal("numbers")
	}
	if r.Bools(&x, &y, &z); !x || y || !z {
		t.Fatal("flags")
	}
	if r.String() != "a" || r.String() != "b" || r.String() != "a" {
		t.Fatal("strings")
	}
	if got := r.Strings(); got != nil {
		t.Fatal("a nil slice must stay nil")
	}
	if got := r.Strings(); got == nil || len(got) != 0 {
		t.Fatal("an empty slice must stay empty")
	}
	if got := r.Strings(); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatal(got)
	}
	if n, isNil := r.Len(); n != 2 || isNil || r.Uint()+r.Uint() != 0 {
		t.Fatal("length")
	}
	table, rest, _ := ReadStrings(b)
	if !r.Done() || r.Err() != nil || r.Offset() != len(rest) {
		t.Fatal("rest")
	}
	// The table holds each string once.
	if !reflect.DeepEqual(table, []string{"a", "b", "c"}) {
		t.Fatal(table)
	}
}

func TestWriterFails(t *testing.T) {
	w := NewWriter(NewStrings())
	w.Uint(1)
	w.Fail(errors.New("first"))
	w.Fail(errors.New("second"))
	w.Uint(2)
	w.Int(3)
	if w.Err().Error() != "first" || len(w.Buf) != 1 {
		t.Fatal("a failed writer writes nothing more and keeps its first error")
	}
	w.Reset()
	if w.Err() != nil || len(w.Buf) != 0 {
		t.Fatal("reset")
	}
	for range MaxDepth {
		if !w.Enter() {
			t.Fatal("within the bound")
		}
	}
	if w.Enter() || w.Err() == nil {
		t.Fatal("past the bound")
	}
	w.Reset()
	w.Enter()
	w.Leave()
	if w.depth != 0 {
		t.Fatal("leave")
	}
}

func TestReaderCorrupt(t *testing.T) {
	cases := map[string]func(r *Reader){
		"uint": func(r *Reader) { r.Uint() },
		"int":  func(r *Reader) { r.Int() },
		// A string past the table.
		"string": func(r *Reader) { _ = r.String() },
		// A length past the input.
		"length": func(r *Reader) { r.Len() },
	}
	inputs := map[string][]byte{"uint": {0x80}, "int": {0x80}, "string": {5}, "length": {9, 1}}
	for name, read := range cases {
		r := NewReader(inputs[name], []string{"a"})
		read(r)
		if !errors.Is(r.Err(), ErrCorrupt) || r.Done() {
			t.Fatalf("%s: not corrupt", name)
		}
		// Every later read gives zero.
		if r.Uint() != 0 || r.Int() != 0 || r.String() != "" || r.Strings() != nil {
			t.Fatalf("%s: read after failure", name)
		}
	}
	r := NewReader(nil, nil)
	for range MaxDepth {
		r.Enter()
	}
	if r.Enter() || !errors.Is(r.Err(), ErrCorrupt) {
		t.Fatal("depth")
	}
	r = NewReader(nil, nil)
	r.Enter()
	r.Leave()
	if r.depth != 0 {
		t.Fatal("leave")
	}
	for name, b := range map[string][]byte{
		"count":     {0x80},
		"too many":  {9},
		"size":      {1, 0x80},
		"too large": {1, 5, 'a'},
	} {
		if _, _, err := ReadStrings(b); !errors.Is(err, ErrCorrupt) {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func BenchmarkWriter(b *testing.B) {
	w := NewWriter(NewStrings())
	b.ReportAllocs()
	for b.Loop() {
		w.Reset()
		for i := range 64 {
			w.Uint(uint64(i))
			w.String("name")
			w.Bools(true, false)
		}
	}
}

// A corrupt length cannot make a small input allocate much: Make and
// MakeMap charge the decoded memory against the input's size and fail the
// reader past it.
func TestAllocationBudget(t *testing.T) {
	r := NewReader(make([]byte, 4), nil)
	if s := Make[int64](r, 8); len(s) != 8 || r.Err() != nil {
		t.Fatalf("within budget: %d %v", len(s), r.Err())
	}
	if s := Make[[64]byte](r, 1000); s != nil || r.Err() == nil {
		t.Fatal("past the budget: allocated")
	}
	if m := MakeMap[string, int](r, 1); m == nil || len(m) != 0 {
		t.Fatal("a failed reader still returns a usable map")
	}
	r = NewReader(make([]byte, 4), nil)
	if m := MakeMap[string, int](r, 2); m == nil || r.Err() != nil {
		t.Fatalf("map within budget: %v", r.Err())
	}
	if m := MakeMap[string, [4096]byte](r, 1000); len(m) != 0 || r.Err() == nil {
		t.Fatal("map past the budget")
	}
	r = NewReader(make([]byte, 1), nil)
	if s := Make[struct{}](r, 64); len(s) != 64 {
		t.Fatal("zero-size elements are charged a byte each")
	}
}
