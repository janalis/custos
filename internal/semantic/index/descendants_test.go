package index

import "testing"

func TestSubclasses(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php
class A {}
class B extends A {}
interface I {}
class X implements I {}
`))
	ix := New(base)
	ix.Add(extract(t, "c.php", `<?php
class C extends B {}
class D extends C {}
class E extends A implements I {}
`))
	got := map[string]bool{}
	for _, s := range ix.Subclasses(`\A`) {
		got[s] = true
	}
	for _, want := range []string{"B", "C", "D", "E"} {
		if !got[want] {
			t.Fatalf("missing %s in %v", want, got)
		}
	}
	if len(got) != 4 {
		t.Fatalf("unexpected: %v", got)
	}
	if s := ix.Subclasses("I"); len(s) != 0 {
		t.Fatalf("interfaces followed: %v", s)
	}
}
