package flow

import "testing"

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
