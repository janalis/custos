package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

// functionNamed returns the declaration of a top-level function.
func functionNamed(f *syntax.File, name string) *syntax.Function {
	for _, s := range f.Stmts {
		if fn, ok := s.(*syntax.Function); ok && fn.Name.Value == name {
			return fn
		}
	}
	return nil
}

// A callable of the file is summarized before the scope that reads it, so
// the summaries of its callers already hold what it returns, and the
// environment then read what the snapshot with those summaries reads.
func TestSummariesServeTheCallablesOfTheFile(t *testing.T) {
	e, f := environment(t, `function a($c){return b($c);} function b($c){return $_GET['x'];} class K { function m(){ return $this->n(); } function n(){ return 1; } }`, nil)
	env, own := Summarize(f, e.types, nil)
	if !tainted(own.Summaries["a"].Return) {
		t.Fatalf("a=%+v", own.Summaries["a"])
	}
	if read := env.answers["b"]; !read.own || read.held || !read.ok {
		t.Fatalf("b answered %+v", read)
	}
	if read := env.answers[sinkClassKey("K")]; !read.own || !read.ok {
		t.Fatalf("class answered %+v", read)
	}
	overlay := (*Snapshot)(nil).WithFile(own)
	if rebased := env.Rebase(overlay); rebased != env || env.snapshot != overlay {
		t.Fatal("environment not kept")
	}
}

// When a callable reaches itself through another, the one its own summary
// reads first has no summary yet: a new environment keeps only the results
// that read nothing the file's summaries change. A caller read the summary
// it was given, not what computing that summary read.
func TestRebaseKeepsTheResultsTheSummariesDoNotChange(t *testing.T) {
	e, f := environment(t, `function a($c){ if ($c) { return b($c); } return $_GET['x']; } function b($c){ return a($c); } function c($x){ return strlen($x); }`, nil)
	env, own := Summarize(f, e.types, nil)
	if read := env.answers["a"]; !read.held || read.ok {
		t.Fatalf("a answered %+v", read)
	}
	rebased := env.Rebase((*Snapshot)(nil).WithFile(own))
	if rebased == env {
		t.Fatal("environment kept with a changed contract")
	}
	for _, name := range []string{"a", "c"} {
		if rebased.scopes[functionNamed(f, name)] != env.scopes[functionNamed(f, name)] {
			t.Errorf("%s not kept", name)
		}
	}
	if _, kept := rebased.scopes[functionNamed(f, "b")]; kept {
		t.Error("b kept")
	}
	// Other runtime assumptions keep nothing.
	assuming := NewSnapshotWithOptions(Options{Assertions: true}, own)
	if other := env.Rebase(assuming); other == env || len(other.scopes) != 0 {
		t.Fatalf("kept %d results under other options", len(other.scopes))
	}
	again, _ := Summarize(f, e.types, nil)
	if other := again.Rebase(NewSnapshotWithOptions(Options{Assertions: true})); other == again || len(other.scopes) != 0 {
		t.Fatalf("kept %d results under other options", len(other.scopes))
	}
}

// The file's summaries answer for a name only where no other file of the
// snapshot declares it.
func TestOnlyACallableNoOtherFileDeclaresIsServed(t *testing.T) {
	file := func(path string, names ...string) *File {
		f := &File{Path: path, Summaries: map[string]Summary{}}
		for _, name := range names {
			f.Summaries[name] = Summary{}
		}
		return f
	}
	shared := NewSnapshot(file("a.php", "x", "y"), file("b.php", "y"), file("c.php", "y"))
	for _, tc := range []struct {
		name     string
		snapshot *Snapshot
		key      string
		alone    bool
	}{
		{"no snapshot", nil, "x", true},
		{"declared elsewhere", shared, "x", false},
		{"declared by its earlier copy", shared, "x", true},
		{"declared by many", shared, "y", false},
		{"undeclared", shared, "z", true},
		{"overlay", shared.WithFile(file("d.php")), "z", false},
	} {
		path := "t.php"
		if tc.name == "declared by its earlier copy" {
			path = "a.php"
		}
		if got := tc.snapshot.alone(path, tc.key); got != tc.alone {
			t.Errorf("%s: alone=%v", tc.name, got)
		}
	}
	e, f := environment(t, `function x(){ return 1; } function w(){ return x(); }`, nil)
	env, _ := Summarize(f, e.types, shared)
	if read := env.answers["x"]; read.own || !read.held {
		t.Fatalf("x answered %+v", read)
	}
}
