package stubs

import (
	"sync"
	"testing"
	"time"

	"custos/internal/phpver"
)

func TestBuiltins(t *testing.T) {
	start := time.Now()
	ix := Index()
	t.Logf("decoded in %v", time.Since(start))
	if f := ix.Function("strlen", phpver.PHP84); f == nil || f.Return != "int" {
		t.Fatalf("strlen: %+v", f)
	}
	if f := ix.Function("str_contains", phpver.PHP84); f == nil {
		t.Fatal("str_contains missing")
	}
	if !ix.IsSubtype("ArrayIterator", "Traversable", phpver.PHP84) {
		t.Fatal("ArrayIterator should implement Traversable")
	}
	if m := ix.FindMethod("DateTime", "format", phpver.PHP84); m == nil {
		t.Fatal("DateTime::format missing")
	}
	if c := ix.Constant("PHP_INT_MAX", 0); c == nil {
		t.Fatal("PHP_INT_MAX missing")
	}
}

func BenchmarkDecode(b *testing.B) {
	for b.Loop() {
		once = *new(sync.Once)
		Index()
	}
}
