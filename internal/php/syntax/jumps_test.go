package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestTailOf(t *testing.T) {
	f := Parse("t.php", []byte(`<?php
namespace N {
	function f($c) {
		before: a();
		return 1;
		one: b();
		if ($c) { two: c(); return 2; }
		function g() { inner: d(); }
		$h = function () { closed: e(); };
		$k = fn () => 1;
		new class { function m() { deep: f(); } };
		three: g();
	}
	function live($c) { only: a(); }
	switch (1) { case 1: return; four: h(); }
	return;
	five: i();
}`), Options{Version: phpversion.Max})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	SetParents(f)
	space := f.Stmts[0].(*Namespace)
	body := space.Stmts[0].(*Function).Body
	tail := TailOf(body)
	if got := labelNames(tail); got != "one two three" {
		t.Fatalf("labels of the function = %q", got)
	}
	if TailOf(body) != tail {
		t.Fatal("the tail of a list is read once")
	}
	for i, want := range map[int]int{2: 2, 3: 2, len(body.Stmts) - 1: 2} {
		if got := tail.LastTerminating(i); got != want {
			t.Errorf("LastTerminating(%d) = %d, want %d", i, got, want)
		}
	}
	if TailOf(space.Stmts[1].(*Function).Body) != nil {
		t.Error("a list no statement ends has no tail")
	}
	if TailOf(space.Stmts[0]) != nil {
		t.Error("a function owns no list")
	}
	if got := labelNames(TailOf(space.Stmts[2].(*Switch).Cases[0])); got != "four" {
		t.Errorf("labels of the case = %q", got)
	}
	if got := labelNames(TailOf(space)); got != "five" {
		t.Errorf("labels of the namespace = %q", got)
	}
	if got := TailOf(space).LastTerminating(5); got != 3 {
		t.Errorf("LastTerminating in the namespace = %d", got)
	}
}

func labelNames(t *Tail) string {
	out := ""
	for i, l := range t.Labels {
		if i > 0 {
			out += " "
		}
		out += l.Name.Value
	}
	return out
}

func TestScopeGotos(t *testing.T) {
	f := Parse("t.php", []byte(`<?php
function f($c) {
	if ($c) { goto a; }
	$h = function () { goto closed; closed: };
	function g() { goto inner; inner: }
	a: goto b;
	b:
}
function none() { return 1; }
abstract class C { abstract function m(); public $p { get => 1; } }
$arrow = fn () => 1;
`), Options{Version: phpversion.Max})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	SetParents(f)
	all := ScopeGotos(f.Stmts[0])
	if len(all) != 2 || all[0].Label.Value != "a" || all[1].Label.Value != "b" {
		t.Fatalf("jumps of f = %v", all)
	}
	if again := ScopeGotos(f.Stmts[0]); len(again) != 2 || again[0] != all[0] {
		t.Fatal("the jumps of a scope are read once")
	}
	if got := ScopeGotos(f.Stmts[1]); len(got) != 0 {
		t.Errorf("jumps of none = %v", got)
	}
	class := f.Stmts[2].(*ClassLike)
	InspectFile(f, func(n Node) bool {
		switch n.(type) {
		case *Method, *PropertyHook, *ArrowFunction:
			if got := ScopeGotos(n); got != nil {
				t.Errorf("%T: jumps = %v", n, got)
			}
		}
		return true
	})
	if ScopeGotos(nil) != nil || ScopeGotos(class) != nil {
		t.Error("no scope, no jumps")
	}
}

func TestScopeLabels(t *testing.T) {
	f := Parse("t.php", []byte(`<?php
function f($c) {
	a: if ($c) { b: goto a; }
	$h = function () { closed: };
	function g() { inner: }
	new class { function m() { deep: } };
	c:
}
function none() { return 1; }
$arrow = fn () => 1;
`), Options{Version: phpversion.Max})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	SetParents(f)
	names := ""
	for _, label := range ScopeLabels(f.Stmts[0]) {
		names += label.Name.Value
	}
	if names != "abc" {
		t.Fatalf("labels of f = %q", names)
	}
	if got := ScopeLabels(f.Stmts[1]); len(got) != 0 {
		t.Errorf("labels of none = %v", got)
	}
	if ScopeLabels(nil) != nil || ScopeLabels(f.Stmts[2].(*ExprStmt).Expr.(*Assign).Value) != nil {
		t.Error("no block body, no labels")
	}
}

func TestScopeMemo(t *testing.T) {
	f := Parse("t.php", []byte(`<?php
function f() { return 1; }
function g() { return 2; }
$arrow = fn () => 1;
`), Options{Version: phpversion.Max})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	SetParents(f)
	built := 0
	build := func(scope Node) *string {
		built++
		name := scope.(*Function).Name.Value
		return &name
	}
	first := ScopeMemo(f.Stmts[0], build)
	if again := ScopeMemo(f.Stmts[0], build); again != first || *first != "f" || built != 1 {
		t.Fatalf("made %d times: %q", built, *first)
	}
	if other := ScopeMemo(f.Stmts[1], build); *other != "g" || built != 2 {
		t.Fatalf("the memo of g = %q, made %d times", *other, built)
	}
	if ScopeMemo(nil, build) != nil || ScopeMemo(f.Stmts[2].(*ExprStmt).Expr.(*Assign).Value, build) != nil || built != 2 {
		t.Error("no block body, no memo")
	}
}
