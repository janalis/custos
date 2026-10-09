package index

import (
	"reflect"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestMagicMethodSignatures(t *testing.T) {
	fs := extract(t, "magic.php", `<?php
namespace App;
use Domain\Item as Thing;
/**
 * @phpstan-type Row array{item: Thing}
 * @method static Thing lookup(Thing $item, Row $row, callable(Thing, int): Thing $callback, string $label = 'a, b', array &$out = ['k' => [1, 2]], string &...$rest)
 * @method omitted($value)
 * @method int real(string &$wrong)
 * @method string broken(array<int] $x)
 * @method void duplicate(int $first)
 * @method void duplicate(string $second)
 */
class Magic { public function real(int $actual): bool {} }
`)
	c := fs.Classes[0]
	m := c.Methods["lookup"]
	if m == nil || !m.Magic || !m.Static || m.DocReturn != `\Domain\Item` {
		t.Fatalf("lookup = %+v", m)
	}
	want := []Param{
		{Name: "item", DocType: `\Domain\Item`},
		{Name: "row", DocType: `array{item:\Domain\Item}`},
		{Name: "callback", DocType: `callable(): (\Domain\Item)`},
		{Name: "label", DocType: "string", Optional: true, Default: "'a, b'"},
		{Name: "out", DocType: "array", Optional: true, ByRef: true, Default: "['k' => [1, 2]]"},
		{Name: "rest", DocType: "string", Optional: true, ByRef: true, Variadic: true},
	}
	if !reflect.DeepEqual(m.Params, want) {
		t.Fatalf("params = %+v; want %+v", m.Params, want)
	}
	if m := c.Methods["omitted"]; m == nil || m.DocReturn != "" || len(m.Params) != 1 || m.Params[0].DocType != "" {
		t.Fatalf("omitted = %+v", m)
	}
	if m := c.Methods["real"]; m.Magic || m.Return != "bool" || len(m.Params) != 1 || m.Params[0].Name != "actual" {
		t.Fatalf("real = %+v", m)
	}
	if c.Methods["broken"] != nil || c.Methods["duplicate"].Params[0].Name != "first" {
		t.Fatal("malformed signature or duplicate declaration changed lookup")
	}
}

func BenchmarkExtractMagicMethodSignatures(b *testing.B) {
	f := syntax.Parse("magic.php", []byte(`<?php
namespace App;
use Domain\Item;
/**
 * @method static Item load(array{item: Item, tags: list<string>} $row, callable(Item, int): Item $callback, array &$out = ['k' => [1, 2]], string &...$rest)
 * @method void notify(string $message = 'a, b')
 */
class Magic {}
`), syntax.Options{Version: phpversion.PHP84})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		Extract(f)
	}
}
