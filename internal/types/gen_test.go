package types

import "testing"

func TestGenericArgs(t *testing.T) {
	c := FromDoc(`Collection<int, Foo>|null`, nil)
	if c.String() != `\Collection|null` || !c.Has(`\Collection`) || len(c.Classes()) != 1 {
		t.Fatalf("atoms must be unchanged: %s", c)
	}
	args := c.TypeArgs(`\Collection`)
	if len(args) != 2 || args[0].String() != "int" || args[1].String() != `\Foo` {
		t.Fatalf("args: %v", args)
	}
	if got := c.DocString(); got != `\Collection<int, \Foo>|null` {
		t.Fatalf("DocString: %s", got)
	}
	if back := FromDoc(c.DocString(), nil); back.ShapeString() != c.ShapeString() {
		t.Fatalf("round trip: %s vs %s", back.ShapeString(), c.ShapeString())
	}
	if w := c.Without("null"); w.TypeArgs(`\Collection`) == nil {
		t.Fatal("Without keeps the arguments")
	}
	if !c.Equal(Of(`\Collection`, "null")) {
		t.Fatal("Equal ignores arguments")
	}
	same := Union(c, FromDoc(`Collection<int, Foo>`, nil))
	if same.TypeArgs(`\Collection`) == nil {
		t.Fatal("union of equal arguments keeps them")
	}
	if Union(c, FromDoc(`Collection<int, Bar>`, nil)).TypeArgs(`\Collection`) != nil {
		t.Fatal("union of different arguments drops them")
	}
	if Union(c, Of(`\Collection`)).TypeArgs(`\Collection`) != nil {
		t.Fatal("union with a bare member drops them")
	}
	if FromDoc(`Collection<Foo>|Collection<Bar>`, nil).TypeArgs(`\Collection`) != nil {
		t.Fatal("conflicting doc members drop them")
	}
	it := FromDoc(`iterable<string, Foo>`, nil)
	if it.String() != "iterable" || len(it.TypeArgs("iterable")) != 2 {
		t.Fatalf("iterable: %s", it.ShapeString())
	}
	if got := FromDoc(`?Box<array{a: int}>`, nil).ShapeString(); got != `\Box<array{a:int}>|null` {
		t.Fatalf("nested shape argument: %s", got)
	}
}

// BenchmarkFromDoc covers doc type parsing (index extraction and inline
// @var): plain unions, generics and shapes.
func BenchmarkFromDoc(b *testing.B) {
	for name, in := range map[string]string{
		"plain":   `int|string|null`,
		"generic": `Collection<int, Foo>|Foo[]`,
		"shape":   `array{id: int, tags: list<string>, meta?: array<string, mixed>}`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = FromDoc(in, nil)
			}
		})
	}
}
