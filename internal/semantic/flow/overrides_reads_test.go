package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

// A second reader of a cached overrides answer depends on the subclass
// contracts the first one read, as the reader that computed it does.
func TestOverridesReplayReads(t *testing.T) {
	e, _ := environment(t, `class A { function m() {} } class B extends A { function m() {} }`, nil)
	e.consulted = map[string]bool{}
	for i := range 2 {
		reads := e.begin()
		if !e.overrides("A", "m", "A") {
			t.Fatal("B::m replaces A::m")
		}
		e.end(reads)
		if !reads["b::m"] {
			t.Fatalf("reader %d: reads=%v, want b::m", i, reads)
		}
	}
}

// A second read of a property whose answer is cached depends on the __get
// of the subclasses, as the first one does.
func TestBackedReplayReads(t *testing.T) {
	e, file := environment(t, `class C { private $p; function a() { return $this->p; } function b() { return $this->p; } } class D extends C { function __get($n) { return 1; } }`, nil)
	e.consulted = map[string]bool{}
	var fetches []*syntax.PropertyFetch
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if f, ok := n.(*syntax.PropertyFetch); ok {
			fetches = append(fetches, f)
		}
		return true
	})
	if len(fetches) != 2 {
		t.Fatalf("fetches: %d", len(fetches))
	}
	for i, fetch := range fetches {
		reads := e.begin()
		if e.backed(fetch) {
			t.Fatal("D::__get may answer $this->p")
		}
		e.end(reads)
		if !reads["d::__get"] {
			t.Fatalf("reader %d: reads=%v, want d::__get", i, reads)
		}
	}
}
