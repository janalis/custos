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

func TestVersionedBuiltinMetadata(t *testing.T) {
	ix := Index()
	if method := ix.FindMethod("DateTimeImmutable", "add", phpversion.PHP85); method == nil || !method.NoDiscard {
		t.Fatal("DateTimeImmutable::add NoDiscard metadata")
	}
	if !ix.Function("is_real", phpversion.PHP74).Avail.In(phpversion.PHP74) || ix.Function("is_real", phpversion.PHP80).Avail.In(phpversion.PHP80) {
		t.Fatal("is_real availability")
	}
	if ix.FindMethod("WeakMap", "offsetGet", phpversion.PHP74) != nil || ix.FindMethod("WeakMap", "offsetGet", phpversion.PHP80) == nil {
		t.Fatal("WeakMap method availability")
	}
	if ix.IsSubtype("DatePeriod", "IteratorAggregate", phpversion.PHP74) || !ix.IsSubtype("DatePeriod", "IteratorAggregate", phpversion.PHP80) {
		t.Fatal("DatePeriod interfaces")
	}
	for _, tc := range []struct {
		v phpversion.Version
		n int
	}{{phpversion.PHP74, 4}, {phpversion.PHP80, 3}} {
		m := ix.FindMethod("PDO", "query", tc.v)
		if m == nil || len(m.Params) != tc.n || !m.Builtin {
			t.Fatalf("PDO query %v: %+v", tc.v, m)
		}
	}
}

func TestUnrestrictedBuiltinMethods(t *testing.T) {
	ix := Index()
	if method := ix.FindMethod("IntlCalendar", "get", 0); method == nil || !method.Builtin {
		t.Fatalf("unrestricted Intl method: %+v", method)
	}
	if method := ix.FindMethod("PDO", "query", 0); method == nil || len(method.Params) != 3 {
		t.Fatalf("unrestricted versioned method: %+v", method)
	}
}
