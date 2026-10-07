package util

import (
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestTestContext(t *testing.T) {
	for p, want := range map[string]bool{
		"src/FooTest.php": true, "a/Spec.php": true, "x.phpt": true, "a/Fixtures/b.php": true,
		"src/Foo.php": false, "a/fixtures/b.php": false,
	} {
		if got := IsTestPath(p); got != want {
			t.Errorf("IsTestPath(%q) = %v", p, got)
		}
	}
	for fqn, want := range map[string]bool{
		`App\FooTest`: true, `\App\Tests\Foo`: true, `App\Test\Foo`: true, `Foo`: false, `App\Testing\Foo`: false,
	} {
		if got := IsTestClassFQN(fqn); got != want {
			t.Errorf("IsTestClassFQN(%q) = %v", fqn, got)
		}
	}
}

func TestChainLookups(t *testing.T) {
	src := `<?php
namespace N;
trait T { private $t; public function tm() {} }
interface I { public function im(); }
/** @property int $magic
 *  @method void mm() */
class A { use T; protected $a; public function am() { return 1; } }
class B extends A implements I { public function im() {} }
class C extends B {}
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	r := names.New(f)
	if p := PropertyInChain(ix, `N\C`, "t", 0); p == nil || p.Class != `N\T` {
		t.Fatalf("trait property: %+v", p)
	}
	if p := PropertyInChain(ix, `N\C`, "a", 0); p == nil || p.Class != `N\A` {
		t.Fatalf("parent property: %+v", p)
	}
	if p := PropertyInChain(ix, `N\C`, "magic", 0); p != nil {
		t.Fatalf("magic property found")
	}
	if m := MethodInChain(ix, `N\C`, "AM", 0); m == nil || m.Class != `N\A` {
		t.Fatalf("method: %+v", m)
	}
	if m := MethodInChain(ix, `N\C`, "mm", 0); m != nil {
		t.Fatalf("magic method found")
	}
	if m := MethodInChain(ix, `N\C`, "tm", 0); m == nil || m.Class != `N\T` {
		t.Fatalf("trait method: %+v", m)
	}
	if m := MethodInChain(ix, `N\C`, "count", 0); m != nil {
		t.Fatalf("unexpected method")
	}
	var decl *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ClassLike); ok && c.Name.Value == "C" {
			decl = c
		}
		return true
	})
	if got := ClassDeclFQN(r, decl); got != `N\C` {
		t.Fatalf("ClassDeclFQN = %q", got)
	}
	if got := ParentFQN(r, decl); got != `N\B` {
		t.Fatalf("ParentFQN = %q", got)
	}
	if ClassDecl(f, ix.Class(`N\C`, 0)) != decl {
		t.Fatalf("ClassDecl mismatch")
	}
	am := MethodInChain(ix, `N\B`, "am", 0)
	if md := MethodDecl(f, ix, am, 0); md == nil || md.Name.Value != "am" {
		t.Fatalf("MethodDecl: %v", md)
	}
	if m := MethodInChain(ix, `ArrayIterator`, "count", 0); m == nil {
		t.Fatalf("stub method not found (spans lost?)")
	}
}
