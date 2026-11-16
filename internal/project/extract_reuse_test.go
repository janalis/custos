package project

import (
	"fmt"
	"reflect"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/infer"
)

// dependencyDecls are typical dependency declarations, numbered by %[1]d,
// whose return types are partly inferred from bodies.
const dependencyDecls = `
/** A collection of values. */
class Bag%[1]d implements \Countable {
	private array $items = [];
	public const LIMIT = 10;
	public function __construct(array $items = []) { $this->items = $items; }
	public function add($value) { $this->items[] = $value; return $this; }
	public function count(): int { return count($this->items); }
	public function first() { return $this->items[0] ?? null; }
	public function names() { return array_map(fn ($v) => (string) $v, $this->items); }
	protected function label() { return 'bag:' . self::LIMIT; }
}

function bag%[1]d(...$values) { return new Bag%[1]d($values); }
`

// dependencySource declares n copies of dependencyDecls.
func dependencySource(n int) []byte {
	src := []byte("<?php\nnamespace Vendor\\Package;\n")
	for i := range n {
		src = fmt.Appendf(src, dependencyDecls, i)
	}
	return src
}

func TestExtractSymbolsAfterRelease(t *testing.T) {
	opt := syntax.Options{Version: phpversion.PHP83, CopyText: true}
	src := dependencySource(1)
	first := ExtractSymbols("bag.php", src, opt, infer.Options{})
	// The second extraction parses into the memory the first one released.
	second := ExtractSymbols("bag.php", src, opt, infer.Options{})
	if !reflect.DeepEqual(first, second) {
		t.Fatal("extraction on recycled parser memory changed the symbols")
	}
	if first.Classes[0].Methods["label"].Inferred != "string" {
		t.Fatal("return type of an untyped method was not inferred")
	}
}

func BenchmarkExtractSymbols(b *testing.B) {
	src := dependencySource(20)
	opt := syntax.Options{Version: phpversion.PHP83, CopyText: true}
	ExtractSymbols("bags.php", src, opt, infer.Options{}) // warm embedded stubs
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		ExtractSymbols("bags.php", src, opt, infer.Options{})
	}
}
