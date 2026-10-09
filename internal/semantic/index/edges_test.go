package index

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestRemoveKeepsOtherDeclarations(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php class A {} function f() {} const K = 1;`))
	ix.Add(extract(t, "b.php", `<?php class A {} function f() {} const K = 2;`))
	if files, classes, funcs, consts := ix.Stats(); files != 2 || classes != 2 || funcs != 2 || consts != 2 {
		t.Fatalf("stats: %d %d %d %d", files, classes, funcs, consts)
	}
	ix.Add(extract(t, "a.php", `<?php`)) // replaces a.php
	if files, classes, funcs, consts := ix.Stats(); files != 2 || classes != 1 || funcs != 1 || consts != 1 {
		t.Fatalf("stats after replace: %d %d %d %d", files, classes, funcs, consts)
	}
	if c := ix.Constant("K", 0); c == nil || c.Value != "2" || c.File != "b.php" {
		t.Fatalf("constant: %+v", c)
	}
}

func TestDropStaleInferredMembers(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "ns.php", `<?php namespace App; function strlen($s) { return 1; }`))
	fs := extract(t, "c.php", `<?php class C { public $p; public function m() {} }`)
	c := fs.Classes[0]
	c.Methods["m"].Inferred, c.Props["p"].Inferred = "int", "int"
	fs.ReturnDeps = []string{`App\other`}
	ix.DropStaleInferred(fs)
	if c.Methods["m"].Inferred != "int" {
		t.Fatal("not stale: kept")
	}
	fs.ReturnDeps = []string{`App\strlen`}
	ix.DropStaleInferred(fs)
	if c.Methods["m"].Inferred != "" || c.Props["p"].Inferred != "" {
		t.Fatal("stale: members cleared")
	}
}

func TestDocEdges(t *testing.T) {
	fs := extract(t, "d.php", `<?php
/**
 * @phpstan-type Id int
 * @template T
 * @extends Plain
 * @extends A<int>|B<int>
 * @implements I<int
 * @method noParens
 * @method ()
 * @method int ok()
 */
class C extends Plain implements I {
    /**
     * @phpstan-assert Id $a
     * @phpstan-assert int
     * @phpstan-assert int $this->m()
     * @phpstan-assert int $this->p->q
     * @phpstan-assert int $this->p[0]
     * @phpstan-assert int $this[0]
     * @phpstan-assert int $a->p
     * @phpstan-assert int $a[0]
     * @phpstan-assert int $missing
     * @phpstan-assert int $this->p
     * @phpstan-assert int $this
     */
    public function check($a) {}
}
`)
	c := fs.Classes[0]
	if len(c.Supers) != 0 {
		t.Errorf("super args: %+v", c.Supers)
	}
	if len(c.Methods) != 2 || c.Methods["ok"] == nil {
		t.Errorf("magic methods: %v", c.Methods)
	}
	as := c.Methods["check"].Asserts
	if len(as) != 3 || as[0].Param != 0 || as[0].Type != "int" || as[1].Param != AssertThisProp || as[1].Prop != "p" || as[2].Param != AssertThis {
		t.Errorf("asserts: %+v", as)
	}
}

func TestStubAttributes(t *testing.T) {
	fs := extract(t, "stub.php", `<?php
#[LanguageLevelTypeAware(['8.0' => 'string'], default: 'int')]
function a() {}
#[LanguageLevelTypeAware(default: 'int')]
function b() {}
#[LanguageLevelTypeAware([], default: '')]
function c() {}
#[LanguageLevelTypeAware]
function d() {}
#[LanguageLevelTypeAware(...)]
#[PhpStormStubsElementAvailable(...)]
function e() {}
#[PhpStormStubsElementAvailable('7.0', '7.4')]
function f() {}
#[PhpStormStubsElementAvailable(from: 'x')]
function g() {}
/**
 * @since 7.1 extra words
 * @removed 8.0
 */
function h() {}
/**
 * @since 5.x
 * @removed 5.3
 */
function i() {}
#[PhpStormStubsElementAvailable(8.1)]
function j() {}
`)
	want := map[string]struct {
		ret      string
		from, to phpversion.Version
	}{
		"a": {"string", 0, 0},
		"b": {"int", 0, 0},
		"c": {"", 0, 0},
		"d": {"", 0, 0},
		"e": {"", 0, 0},
		"f": {"", phpversion.PHP70, phpversion.PHP74},
		"g": {"", 0, 0},
		"h": {"", phpversion.PHP71, phpversion.PHP80 - 1},
		"i": {"", 0, 0},
		"j": {"", phpversion.PHP81, 0}, // unquoted version
	}
	if len(fs.Functions) != len(want) {
		t.Fatalf("got %d functions", len(fs.Functions))
	}
	for _, f := range fs.Functions {
		w := want[f.FQN]
		if f.Return != w.ret || f.Avail.From != w.from || f.Avail.To != w.to {
			t.Errorf("%s: return %q avail %v-%v; want %q %v-%v", f.FQN, f.Return, f.Avail.From, f.Avail.To, w.ret, w.from, w.to)
		}
	}
}

func TestDefineMalformed(t *testing.T) {
	fs := extract(t, "d.php", `<?php define('A', ...); define(B, 1); define('C'); \define('D', 1);`)
	if len(fs.Constants) != 1 || fs.Constants[0].FQN != "D" {
		t.Fatalf("constants: %+v", fs.Constants)
	}
}
