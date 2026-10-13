package flowquery

import (
	"strings"
	"testing"
	"time"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/testing/testbudget"
)

func TestReachable(t *testing.T) {
	src := `<?php function g($c) {
		a1();
		if ($c) { return; }
		a2();
		if ($c) { return; } else { throw new E(); }
		a3();
	}
	function h() { while (true) { x(); } a4(); }
	function k() { while (true) { break; } a5(); switch (1) { case 1: return; default: return; } a6(); }
	function m() { try { return; } catch (E $e) { a7(); } a8(); }`
	f := Parse(t, src)
	want := map[string]bool{"a1": true, "a2": true, "a3": false, "a4": false, "a5": true, "a6": false, "a7": true, "a8": true}
	got := map[string]bool{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name := astquery.CallLastName(c); len(name) == 2 && name[0] == 'a' {
				got[name] = Reachable(c, syntax.EnclosingFuncLike(c))
			}
		}
		return true
	})
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

func TestReachableCaseCondition(t *testing.T) {
	f := Parse(t, `<?php function g($x) { return; switch ($x) { case a9(): } }`)
	var call syntax.Node
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			call = c
		}
		return true
	})
	// The case condition is not in the case's statement list: every
	// statement of the enclosing lists before the switch counts.
	if Reachable(call, syntax.EnclosingFuncLike(call)) {
		t.Fatal("case condition after return is unreachable")
	}
}

// A statement written after one that ends its list still runs when a jump
// lands on a label between the two.
func TestReachableAfterALabel(t *testing.T) {
	src := `<?php namespace N {
	function g($c) { if ($c) { goto out; } return; out: b1(); }
	function h($c) { return; dead: b2(); }
	function k($c) { if ($c) { goto in; } return; if ($c) { b3(); in: b4(); } }
	function m($c) { if ($c) { goto late; } return; b5(); late: ; }
	function n($c) { return; $f = function () { goto x; x: ; }; other: b6(); }
	function o($c) { $f = function () { goto there; }; return; there: b7(); }
	goto top; exit; top: b8();
	return; b9(); }`
	f := Parse(t, src)
	want := map[string]bool{"b1": true, "b2": false, "b3": false, "b4": true, "b5": false, "b6": false, "b7": false, "b8": true, "b9": false}
	got := map[string]bool{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name := astquery.CallLastName(c); len(name) == 2 && name[0] == 'b' {
				got[name] = Reachable(c, syntax.EnclosingVariableScope(c))
			}
		}
		return true
	})
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

// A label only brings a statement back when control can go on from it to the
// statement and a jump that can itself be reached names it.
func TestReachableOnlyThroughLabelsThatLead(t *testing.T) {
	src := `<?php
	function a($c) { if ($c) { goto done; } return 1; function inner() { done: return 1; } c1(); done: return 2; }
	function b($c) { if ($c) { goto done; } return 1; class Inner { function m() { done: return 1; } } c2(); done: return 2; }
	function c($c) { if ($c) { goto mid; } return 1; if ($c > 2) { mid: $c++; } else { c3(); } }
	function d($c) { if ($c) { goto mid; } return 1; while ($c) { if ($c > 2) { mid: $c++; } else { c4(); } } }
	function e($c) { return 1; again: $c++; if ($c < 9) { goto again; } c5(); }
	function f($c) { if ($c) { goto mid; } if ($c > 5) { return 1; mid: $c++; } else { return 2; } c6(); }
	function g($c) { if ($c) { goto mid; } { return 1; mid: $c++; } c7(); }
	function h($c) { if ($c) { goto mid; } if ($c > 5) { return 1; mid: return 3; } else { return 2; } c8(); }
	function i($c) { if ($c) { goto mid; } return 1; while ($c) { mid: if ($c > 3) { break; } return 4; } c9(); }
	function j($c) { if ($c) { goto mid; } return 1; try { mid: other(); return 4; } catch (E $e) { } d1(); }
	function k($c) { if ($c) { goto mid; } return 1; switch ($c) { case 1: mid: if ($c > 3) { break; } return 4; } d2(); }
	function l($c) { if ($c) { goto mid; } return 1; if ($c > 2) { if ($c > 3) { mid: $c++; } return 5; } d3(); }
	function m($c) { if ($c) { goto one; } return 1; one: $c++; goto two; return 2; two: d4(); }
	function n($c) { return 1; one: $c++; goto two; two: $c--; goto one; d5(); }
	function o($c) { if ($c) { goto mid; } return 1; if ($c > 2) { mid: $c++; if ($c > 3) { d6(); } } else { d7(); } }
	function p($c) { if ($c) { goto mid; } return 1; mid: $c++; return 2; d8(); }
	`
	f := Parse(t, src)
	want := map[string]bool{
		"c1": false, "c2": false, "c3": false, "c4": true, "c5": false, "c6": true, "c7": true, "c8": false, "c9": true,
		"d1": true, "d2": true, "d3": false, "d4": true, "d5": false, "d6": true, "d7": false, "d8": false,
	}
	got := map[string]bool{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name := astquery.CallLastName(c); len(name) == 2 && (name[0] == 'c' || name[0] == 'd') {
				got[name] = Reachable(c, syntax.EnclosingVariableScope(c))
			}
		}
		return true
	})
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

// Past a number of jumps whose own reachability would have to be asked, a
// jump counts as taken.
func TestReachableStopsAskingAboutJumps(t *testing.T) {
	for jumps, want := range map[int]bool{jumpChecks: false, jumpChecks + 1: true} {
		f := Parse(t, `<?php function f($c) { return 1; again: probe(); `+strings.Repeat(`goto again; `, jumps)+`}`)
		var call *syntax.FuncCall
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				call = c
			}
			return true
		})
		if got := Reachable(call, syntax.EnclosingVariableScope(call)); got != want {
			t.Errorf("%d jumps: Reachable = %v, want %v", jumps, got, want)
		}
	}
}

// deadTail is a function whose n statement pairs follow a `return`, with a
// label a live jump names at the head of them when labelled.
func deadTail(n int, labelled bool) string {
	var b strings.Builder
	b.WriteString("<?php function f(int $n) { ")
	if labelled {
		b.WriteString("if ($n) { goto L; } return 0; L: ")
	} else {
		b.WriteString("return 0; ")
	}
	for i := 0; i < n; i++ {
		b.WriteString("$x = $n % 3; probe($x);\n")
	}
	b.WriteString("}")
	return b.String()
}

// probesOf returns the calls of a file.
func probesOf(f *syntax.File) (calls []*syntax.FuncCall) {
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			calls = append(calls, c)
		}
		return true
	})
	return calls
}

// reachableCount asks Reachable of every call and counts the reachable ones.
func reachableCount(calls []*syntax.FuncCall) (count int) {
	for _, c := range calls {
		if Reachable(c, syntax.EnclosingVariableScope(c)) {
			count++
		}
	}
	return count
}

// TestReachableLongDeadTailLinear guards the cached tails: asking Reachable
// of every node written after a `return` must stay linear in their number,
// with and without a label among them (it walked back to the `return` for
// each node: 46 s for 8,000 statements).
func TestReachableLongDeadTailLinear(t *testing.T) {
	for labelled, want := range map[bool]int{false: 0, true: 20000} {
		f := Parse(t, deadTail(20000, labelled))
		calls := probesOf(f)
		start := time.Now()
		got := reachableCount(calls)
		elapsed := time.Since(start)
		if got != want {
			t.Errorf("labelled=%v: %d reachable calls, want %d", labelled, got, want)
		}
		if limit := testbudget.Of(2 * time.Second); elapsed > limit {
			t.Errorf("labelled=%v: %v for 20,000 dead statements (limit %v)", labelled, elapsed, limit)
		}
	}
}

// BenchmarkReachableDeadTail measures Reachable on every node of a long run
// of statements after a `return`, without and with a label that brings them
// back.
func BenchmarkReachableDeadTail(b *testing.B) {
	for name, labelled := range map[string]bool{"plain": false, "labelled": true} {
		b.Run(name, func(b *testing.B) {
			src := []byte(deadTail(2000, labelled))
			b.ReportAllocs()
			for b.Loop() {
				// Parsed anew each time: the tails are cached on the tree.
				b.StopTimer()
				calls := probesOf(syntax.Parse("b.php", src, syntax.Options{Version: phpversion.Max}))
				b.StartTimer()
				reachableCount(calls)
			}
		})
	}
}
