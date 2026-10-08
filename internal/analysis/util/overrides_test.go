package util

import (
	"testing"

	"custos/internal/index"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestDescendantMethods(t *testing.T) {
	src := `<?php
namespace N;
interface I { public function run(); }
class A implements I { public function run() {} public function name() {} public function only() {} }
class B extends A { public function name() {} }
class C extends B { public function run() {} }
class D extends A {}
class E extends D {}
interface K extends I {}
class G implements I, K {}
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	for _, c := range []struct {
		fqn, name string
		want      bool
	}{
		{`N\A`, "name", true},
		{`N\A`, "run", true},
		{`\N\I`, "run", true},
		{`N\A`, "only", false},
		{`N\B`, "name", false},
		{`N\D`, "run", false},
		{`N\Missing`, "run", false},
		{`N\I`, "absent", false},
	} {
		if got := DescendantMethods(ix, c.fqn, 0)[c.name]; got != c.want {
			t.Errorf("DescendantMethods(%s, %s) = %v", c.fqn, c.name, got)
		}
	}
}

func TestDescendantMethodsCycle(t *testing.T) {
	f := syntax.Parse("x.php", []byte(`<?php class A extends B { function m() {} } class B extends A {}`), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !DescendantMethods(ix, "A", 0)["m"] {
		t.Error("a class in a cycle is its own descendant")
	}
}
