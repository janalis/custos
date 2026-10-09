package stubs

import (
	"bytes"
	"compress/gzip"
	"sync"
	"testing"
	"time"

	phpversion "custos/internal/php/version"
)

func TestBuiltins(t *testing.T) {
	start := time.Now()
	ix := Index()
	t.Logf("decoded in %v", time.Since(start))
	if f := ix.Function("strlen", phpversion.PHP84); f == nil || f.Return != "int" {
		t.Fatalf("strlen: %+v", f)
	}
	if f := ix.Function("str_contains", phpversion.PHP84); f == nil {
		t.Fatal("str_contains missing")
	}
	if !ix.IsSubtype("ArrayIterator", "Traversable", phpversion.PHP84) {
		t.Fatal("ArrayIterator should implement Traversable")
	}
	if m := ix.FindMethod("DateTime", "format", phpversion.PHP84); m == nil {
		t.Fatal("DateTime::format missing")
	}
	if c := ix.Constant("PHP_INT_MAX", 0); c == nil || !c.Builtin {
		t.Fatalf("PHP_INT_MAX missing builtin marker: %+v", c)
	}
}

func BenchmarkDecode(b *testing.B) {
	for b.Loop() {
		once = sync.Once{}
		Index()
	}
}

func TestDecodeErrors(t *testing.T) {
	if _, err := decode([]byte("not gzip")); err == nil {
		t.Error("non-gzip data must fail")
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("not gob"))
	_ = zw.Close()
	if _, err := decode(buf.Bytes()); err == nil {
		t.Error("non-gob payload must fail")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("must should panic on error")
		}
	}()
	must(decode([]byte("x")))
}
