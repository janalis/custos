package index

import "testing"

func TestClassDecls(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php namespace App; class User {}`))
	base.Add(extract(t, "b.php", `<?php namespace App; class User {} class Other {}`))
	top := New(base)
	top.Add(extract(t, "a.php", `<?php namespace App; class User {}`))
	if got := len(top.ClassDecls(`\App\User`, 0)); got != 2 {
		t.Fatalf("App\\User: got %d decls, want 2", got)
	}
	if got := len(top.ClassDecls(`app\other`, 0)); got != 1 {
		t.Fatalf("App\\Other: got %d decls, want 1", got)
	}
	if got := len(top.ClassDecls(`App\Missing`, 0)); got != 0 {
		t.Fatalf("missing: got %d", got)
	}
}

func TestChildrenAll(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php class Base {} class A extends Base {}`))
	top := New(base)
	top.Add(extract(t, "b.php", `<?php class B extends Base {} class C extends \Base {}`))
	if got := top.ChildrenAll(`\Base`); len(got) != 3 {
		t.Fatalf("got %v, want 3 children", got)
	}
	if got := top.ChildrenAll(`Missing`); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
