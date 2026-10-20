package heapiterationconsumescollection

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewLocalClone covers iteration over a local that holds a clone: only
// a variable provably bound to a private copy silences the report.
func TestReviewLocalClone(t *testing.T) {
	const pre = "function f(SplMinHeap $h, $c, $n) { "
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"clone local", pre + "$copy = clone $h; foreach ($copy as $v) {} }", 0},
		{"parenthesised clone", pre + "$copy = (clone $h); foreach ($copy as $v) {} }", 0},
		{"top level", "$h = new SplMinHeap(); $copy = clone $h; foreach ($copy as $v) {}", 0},
		{"reads between", pre + "$copy = clone $h; $copy->insert(1); echo count($copy); foreach ($copy as $v) {} foreach ($copy as $v) {} }", 0},
		{"loop around", pre + "$copy = clone $h; while ($c) { foreach ($copy as $v) {} } }", 0},
		{"same block in branch", pre + "if ($c) { $copy = clone $h; foreach ($copy as $v) {} } }", 0},
		{"priority queue", "function f(SplPriorityQueue $h) { $copy = clone $h; foreach ($copy as $v) {} }", 0},

		{"plain alias", pre + "$copy = $h; foreach ($copy as $v) {} }", 1},
		{"clone on one branch", pre + "$copy = $h; if ($c) { $copy = clone $h; } foreach ($copy as $v) {} }", 1},
		{"conditional clone only", "function f(SplMinHeap $copy, $c) { if ($c) { $copy = clone $copy; } foreach ($copy as $v) {} }", 1},
		{"reassigned", pre + "$copy = clone $h; $copy = $h; foreach ($copy as $v) {} }", 1},
		{"reassigned in loop", pre + "$copy = clone $h; while ($c) { foreach ($copy as $v) {} $copy = $h; } }", 1},
		{"two clones", pre + "$copy = clone $h; $copy = clone $h; foreach ($copy as $v) {} }", 1},
		{"clone after loop", pre + "$copy = $h; foreach ($copy as $v) {} $copy = clone $h; }", 1},
		{"reference alias", pre + "$copy = clone $h; $r = &$copy; $r = $h; foreach ($copy as $v) {} }", 1},
		{"reference bind", pre + "$copy = clone $h; $copy = &$h; foreach ($copy as $v) {} }", 1},
		{"compound write", "function f(SplMinHeap $copy, SplMinHeap $h) { $copy ??= clone $h; foreach ($copy as $v) {} }", 1},
		{"nested assignment", pre + "$other = $copy = clone $h; foreach ($copy as $v) {} }", 1},
		{"escapes to call", pre + "$copy = clone $h; swap($copy, $h); foreach ($copy as $v) {} }", 1},
		{"unpacked", pre + "$copy = clone $h; count(...$copy); foreach ($copy as $v) {} }", 1},
		{"user count", "namespace N; function count($x) {} function f(\\SplMinHeap $h) { $copy = clone $h; count($copy); foreach ($copy as $v) {} }", 1},
		{"stored", pre + "$copy = clone $h; $keep = $copy; foreach ($copy as $v) {} }", 1},
		{"loop target", "function f(SplMinHeap $h, SplMinHeap ...$n) { $copy = clone $h; foreach ($n as $copy) {} foreach ($copy as $v) {} }", 1},
		{"method argument", pre + "$copy = clone $h; $n->take($copy); foreach ($copy as $v) {} }", 1},
		{"globals", "$h = new SplMinHeap(); $copy = clone $h; $GLOBALS['copy'] = $h; foreach ($copy as $v) {}", 1},
		{"global in function", "$h = new SplMinHeap(); $copy = clone $h; function g() { global $copy, $h; $copy = $h; } g(); foreach ($copy as $v) {}", 1},
		{"property iterable", "class K { public SplMinHeap $p; function f(SplMinHeap $h) { $this->p = clone $h; foreach ($this->p as $v) {} } }", 1},

		// Type inference already gives up on these variables, so nothing is
		// reported; the proof must reject them on its own all the same.
		{"closure reference", pre + "$copy = clone $h; $g = function () use (&$copy, $h) { $copy = $h; }; $g(); foreach ($copy as $v) {} }", 0},
		{"variable variable", pre + "$copy = clone $h; $$n = $h; foreach ($copy as $v) {} }", 0},
		{"extract", pre + "$copy = clone $h; extract($n); foreach ($copy as $v) {} }", 0},
		{"parse_str", pre + "$copy = clone $h; parse_str($n); foreach ($copy as $v) {} }", 0},
		{"include", pre + "$copy = clone $h; include $n; foreach ($copy as $v) {} }", 0},
		{"eval", pre + "$copy = clone $h; eval($n); foreach ($copy as $v) {} }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HeapIterationConsumesCollection"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
