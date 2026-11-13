package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// unseenOrderEnv parses src and returns a PHPDoc-aware Env with the
// expressions echoed by the file, in source order.
func unseenOrderEnv(t *testing.T, src string) (*Env, []syntax.Expr) {
	t.Helper()
	f := syntax.Parse("order.php", []byte(src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	var echoed []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if echo, ok := n.(*syntax.Echo); ok {
			echoed = append(echoed, echo.Exprs[0])
		}
		return true
	})
	return NewEnv(f, names.New(f), ix, phpversion.PHP84, Options{}), echoed
}

// A literal value checked against unseen writes gives the same type
// whatever was typed before: typing the receiver of the call that takes the
// local (itself a local with a value) must not end the check early, and a
// read typed during the check must not keep a type that depends on it.
func TestUnseenWritesOrderIndependent(t *testing.T) {
	for _, src := range []string{
		`<?php function f() { $id = 5; $a = 1; $id->find($a); echo $a; echo $id; }`,
		`<?php function f() { $id = 5; $a = "ab"; $id->find($a[0]); echo $a; echo $id; }`,
		`<?php function f() { $id = 5; $a = "ab"; $b = 2; $id->find($b); $b->find($a[0]); echo $a; echo $b; echo $id; }`,
		`<?php function f() { $id = 5; $a = 1; $o = $id; $o->find($a); echo $a; echo $o; }`,
		`<?php function f() { $id = 5; $a = 1; $b = 2; $b->find($a); $id->find($b); echo $a; echo $b; echo $id; }`,
		// $r, typed while $a is checked, keeps the definition an unknown
		// callee would make only for that check.
		`<?php function f($x) { $a = "ab"; $r = 2; $x->m($r); $r->find($a[0]); echo $a; echo $r; }`,
		`<?php function f($x) { $a = "ab"; $r = rand() ? null : new ArrayObject(); if ($r === null) { return; } $x->m($r); $r->find($a[0]); echo $a; echo $r; }`,
	} {
		forward, fe := unseenOrderEnv(t, src)
		backward, be := unseenOrderEnv(t, src)
		fwd := make([]string, len(fe))
		for i, x := range fe {
			fwd[i] = forward.TypeOf(x).ShapeString()
		}
		bwd := make([]string, len(be))
		for i := len(be) - 1; i >= 0; i-- {
			bwd[i] = backward.TypeOf(be[i]).ShapeString()
		}
		for i := range fwd {
			if fwd[i] != bwd[i] {
				t.Errorf("%s: echo %d is %s typed first, %s typed last", src, i, fwd[i], bwd[i])
			}
		}
	}
}

// A string read where a write its definitions do not show may store any
// value keeps neither the values it was limited to nor those ruled out.
func TestUnseenWritesDropStringExclusions(t *testing.T) {
	env, echoed := unseenOrderEnv(t, `<?php function f() {
		$s = 'a';
		$r = &$s;
		if ($s !== 'b') { $t = $s; $q = &$t; echo $t; $q = 'b'; echo $t; }
	}`)
	for i, x := range echoed {
		got := env.TypeOf(x)
		if _, listed := got.StringValueList(); listed || got.StringExcludes("b") || !got.Has("string") {
			t.Errorf("echo %d: %s rules values in or out", i, got.ShapeString())
		}
	}
}
