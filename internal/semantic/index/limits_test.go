package index

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/testing/testbudget"
)

// Every member lookup walks the ancestors: they are cached per class and
// version and capped at MaxAncestors (security review 2026-10-07: a 5,000
// class extends chain took 12 s to analyse, a 3,000-class cycle 7 s).
func TestAncestorsBounded(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		// Linear work grows 4x from n to 4n, quadratic work 16x: the ratio is
		// independent of machine load and coverage instrumentation, which a
		// fixed time limit is not.
		small, large := lookupTime(t, 2500, cycle), lookupTime(t, 10000, cycle)
		ratio := float64(large) / float64(small)
		t.Logf("cycle=%v: 2.5k %v, 10k %v (x%.1f)", cycle, small, large, ratio)
		if ratio > 9 && !raceEnabled {
			t.Errorf("cycle=%v: 4x more classes took x%.1f longer (super-linear)", cycle, ratio)
		}
		if large > testbudget.Of(60*time.Second) {
			t.Errorf("cycle=%v: lookups took %v", cycle, large)
		}
	}
}

// lookupTime builds an n-class extends chain (or cycle), checks the cap and
// returns the best of two timings of member lookups on every class.
func lookupTime(t *testing.T, n int, cycle bool) time.Duration {
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
	best := time.Duration(1<<63 - 1)
	for run := 0; run < 2; run++ {
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
		if d := time.Since(start); d < best {
			best = d
		}
		if !cycle && found < MaxAncestors {
			t.Errorf("n=%d: found m from %d classes, want at least %d", n, found, MaxAncestors)
		}
		if got := len(ix.Ancestors(fmt.Sprintf("C%d", n-1), 0)); got != MaxAncestors {
			t.Errorf("n=%d cycle=%v: %d ancestors, want the cap %d", n, cycle, got, MaxAncestors)
		}
	}
	return best
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
