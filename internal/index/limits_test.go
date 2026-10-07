package index

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/syntax"
	"custos/internal/testbudget"
)

// Every member lookup walks the ancestors: they are cached per class and
// version and capped at MaxAncestors (security review 2026-10-07: a 5,000
// class extends chain took 12 s to analyse, a 3,000-class cycle 7 s).
func TestAncestorsBounded(t *testing.T) {
	const n = 20000
	for _, cycle := range []bool{false, true} {
		var b strings.Builder
		b.WriteString("<?php\n")
		for i := 0; i < n; i++ {
			parent := i - 1
			if cycle {
				parent = (i + 1) % n
			}
			if parent < 0 {
				fmt.Fprintf(&b, "class C%d { public function m() {} }\n", i)
			} else {
				fmt.Fprintf(&b, "class C%d extends C%d { public function f%d() {} }\n", i, parent, i)
			}
		}
		f := syntax.Parse("t.php", []byte(b.String()), syntax.Options{})
		ix := New(nil)
		ix.Add(Extract(f))
		start := time.Now()
		found := 0
		for i := 0; i < n; i++ {
			c := fmt.Sprintf("C%d", i)
			if ix.FindMethod(c, "m", 0) != nil {
				found++
			}
			ix.FindProperty(c, "p", 0)
			ix.IsSubtype(c, "C0", 0)
			ix.ParentChain(c, 0)
		}
		d := time.Since(start)
		t.Logf("cycle=%v: %v", cycle, d)
		if d > testbudget.Of(3*time.Second) && !raceEnabled {
			t.Errorf("cycle=%v: lookups took %v", cycle, d)
		}
		if !cycle && found < MaxAncestors {
			t.Errorf("found m from %d classes, want at least %d", found, MaxAncestors)
		}
		if got := len(ix.Ancestors(fmt.Sprintf("C%d", n-1), 0)); got != MaxAncestors {
			t.Errorf("cycle=%v: %d ancestors, want the cap %d", cycle, got, MaxAncestors)
		}
	}
}

// The ancestors cache follows changes to the index and to its base.
func TestAncestorsCacheInvalidation(t *testing.T) {
	parse := func(path, src string) *FileSymbols {
		return Extract(syntax.Parse(path, []byte(src), syntax.Options{}))
	}
	base := New(nil)
	base.Add(parse("b.php", "<?php class B {}"))
	ix := New(base)
	ix.Add(parse("a.php", "<?php class A extends B {}"))
	if got := len(ix.Ancestors("A", 0)); got != 2 {
		t.Fatalf("got %d ancestors", got)
	}
	base.Add(parse("b.php", "<?php class B extends Z {} class Z {}"))
	if got := len(ix.Ancestors("A", 0)); got != 3 {
		t.Errorf("after base change: got %d ancestors, want 3", got)
	}
	ix.Remove("a.php")
	if got := len(ix.Ancestors("A", 0)); got != 0 {
		t.Errorf("after remove: got %d ancestors, want 0", got)
	}
}
