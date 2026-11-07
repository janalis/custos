package index

import (
	"fmt"
	"slices"
	"testing"
)

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

// Walks that stop at the first match must read the same classes on every
// run, whatever order the classes were indexed in.
func TestChildrenSorted(t *testing.T) {
	ix := New(nil)
	for i := range 32 {
		ix.Add(extract(t, fmt.Sprintf("c%d.php", i), fmt.Sprintf("<?php\nclass C%02d extends Base {}\n", 31-i)))
	}
	if got := ix.Children("Base"); len(got) != 32 || !slices.IsSorted(got) {
		t.Fatalf("children: %v", got)
	}
}

func TestExtendingBreadthFirstThroughAliases(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php
class A {}
class B extends A {}
interface I {}
class X implements I {}
`))
	ix := New(base)
	ix.Add(extract(t, "c.php", `<?php
class_alias('A', 'Old');
class C extends B {}
class D extends Old {}
class E extends A implements I {}
class F extends E {}
`))
	want := []string{"E", "B", "D", "F", "C"}
	for _, name := range []string{"A", `\A`, `\A`} {
		if got := ix.Extending(name); !slices.Equal(got, want) {
			t.Fatalf("%s: %v, want %v", name, got, want)
		}
	}
	if got := ix.Extending("I"); !slices.Equal(got, []string{"E", "X", "F"}) {
		t.Fatalf("I: %v", got)
	}
	if got := ix.Extending("F"); len(got) != 0 {
		t.Fatalf("F: %v", got)
	}
}

func TestExtendingCycle(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php
class A extends B {}
class B extends A {}
`))
	if got := ix.Extending("A"); !slices.Equal(got, []string{"B"}) {
		t.Fatal(got)
	}
}

func TestExtendingFollowsChanges(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php
class A {}
class B extends A {}
`))
	if got := ix.Extending("A"); !slices.Equal(got, []string{"B"}) {
		t.Fatal(got)
	}
	ix.Add(extract(t, "c.php", `<?php
class C extends A {}
`))
	if got := ix.Extending("A"); !slices.Equal(got, []string{"B", "C"}) {
		t.Fatal(got)
	}
	ix.Remove("a.php")
	if got := ix.Extending("A"); !slices.Equal(got, []string{"C"}) {
		t.Fatal(got)
	}
}

func BenchmarkExtending(b *testing.B) {
	ix := New(nil)
	fs := &FileSymbols{Path: "h.php"}
	fs.Classes = append(fs.Classes, &Class{FQN: "Root"})
	for i := range 300 {
		fs.Classes = append(fs.Classes, &Class{FQN: fmt.Sprintf("Child%d", i), Parent: "Root"})
		fs.ClassAliases = append(fs.ClassAliases, [2]string{fmt.Sprintf("Alias%d", i), fmt.Sprintf("Other%d", i)})
	}
	ix.Add(fs)
	ix.Freeze()
	b.ReportAllocs()
	for b.Loop() {
		ix.Extending(`\Root`)
	}
}
