package flowquery

import (
	"fmt"
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

// However many jumps name a label, it only counts when one of them can be
// reached: none is taken for granted past a number of them.
func TestReachableAsksAboutEveryJump(t *testing.T) {
	for jumps, want := range map[string]bool{
		strings.Repeat(`goto again; `, 200):                        false,
		strings.Repeat(`goto again; `, 200) + `live: goto again; `: true,
	} {
		f := Parse(t, `<?php function f($c) { if ($c) { goto live; } return 1; again: probe(); `+jumps+`}`)
		call := probesOf(f)[0]
		if got := Reachable(call, syntax.EnclosingVariableScope(call)); got != want {
			t.Errorf("%d bytes of jumps: Reachable = %v, want %v", len(jumps), got, want)
		}
	}
}

// reachableByName asks Reachable of every call of the file, with the scope
// scopeOf gives, and returns the answers by function name.
func reachableByName(f *syntax.File, scopeOf func(syntax.Node) syntax.Node) map[string]bool {
	got := map[string]bool{}
	for _, c := range probesOf(f) {
		got[astquery.CallLastName(c)] = Reachable(c, scopeOf(c))
	}
	return got
}

// An endless loop is not left from a label written in it: what follows the
// loop stays out of reach, unless a `try`, a `switch` or another loop around
// it is left some other way.
func TestReachableNotOutOfAnEndlessLoop(t *testing.T) {
	const head = `if ($c) { goto in; } return 1; `
	want := map[string]bool{}
	src := "<?php\n"
	for i, tc := range []struct {
		body string
		want bool
	}{
		{`for (;;) { in: x(); } e%d();`, false},
		{`while (true) { in: x(); } e%d();`, false},
		{`do { in: x(); } while (true); e%d();`, false},
		{`{ for (;;) { in: x(); } } e%d();`, false},
		{`for (;;) { while ($c) { in: x(); } } e%d();`, false},
		{`for (;;) { switch ($c) { case 1: in: x(); } } e%d();`, false},
		{`for (;;) { try { in: x(); } catch (E $e) { } } e%d();`, false},
		{`while (true) { in: if ($c) { break; } } e%d();`, true},
		{`do { return 2; in: x(); } while ($c); e%d();`, true},
		{`try { for (;;) { in: x(); } } catch (E $e) { } e%d();`, true},
		{`while ($c) { for (;;) { in: x(); } } e%d();`, true},
		{`switch ($c) { case 1: for (;;) { in: x(); } } e%d();`, true},
		{`for (;;) { in: x(); } late: e%d();`, false},
	} {
		name := "e" + string(rune('a'+i))
		src += "function f" + name + "($c) { " + head + strings.ReplaceAll(tc.body, "e%d", name) + " }\n"
		want[name] = tc.want
	}
	got := reachableByName(Parse(t, src), syntax.EnclosingVariableScope)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

// A label in one `catch` does not lead to a later `catch` of the same `try`;
// one in the body of the `try` leads to each of them, and any of them to the
// `finally` block.
func TestReachableAcrossCatchBlocks(t *testing.T) {
	src := `<?php
	function a($c) { if ($c) { goto in; } return 1; try { x(); } catch (A $a) { in: x(); } catch (B $b) { h1(); } }
	function b($c) { if ($c) { goto in; } return 1; try { in: x(); } catch (A $a) { h2(); } catch (B $b) { h3(); } }
	function c($c) { if ($c) { goto in; } return 1; try { x(); } catch (A $a) { in: x(); } finally { h4(); } }
	function d($c) { if ($c) { goto in; } return 1; try { x(); } catch (A $a) { h5(); } catch (B $b) { in: x(); if ($c) { h6(); } } }
	function e($c) { if ($c) { goto in; } return 1; while ($c) { try { x(); } catch (A $a) { in: x(); } catch (B $b) { h7(); } } }
	function f($c) { if ($c) { goto in; } return 1; try { x(); } catch (A $a) { in: x(); } catch (B $b) { x(); } catch (C $c) { if ($c) { h8(); } } }
	namespace N { return; try { x(); } catch (A $a) { top: x(); } catch (B $b) { h9(); } }
	`
	want := map[string]bool{"h1": false, "h2": true, "h3": true, "h4": true, "h5": false, "h6": true, "h7": true, "h8": false, "h9": false}
	got := reachableByName(Parse(t, src), syntax.EnclosingVariableScope)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

// A property hook is entered when its property is read or written: what is
// written around its class does not stop it, also when the scope asked about
// is the function around the class or none, and its labels count by its own
// jumps.
func TestReachableInAPropertyHook(t *testing.T) {
	src := `<?php namespace N {
	return;
	class C { public int $x { get { k1(); return 1; } } }
	function f() { return 1; new class { public int $x { get { k2(); return 1; k3(); } } }; }
	final class P { function __construct(public int $x = 0 { get { if ($this->x) { goto in; } return 1; in: k4(); return 2; dead: k5(); } }) { goto dead; dead: } }
	final class Q { public int $x { set { if ($value) { goto in; } return; in: k6(); return; dead: k7(); } } }
	}`
	want := map[string]bool{"k1": true, "k2": true, "k3": false, "k4": true, "k5": false, "k6": true, "k7": false}
	for name, scopeOf := range map[string]func(syntax.Node) syntax.Node{"function": syntax.EnclosingFuncLike, "variable scope": syntax.EnclosingVariableScope} {
		got := reachableByName(Parse(t, src), scopeOf)
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s: Reachable(%s) = %v, want %v", name, k, got[k], v)
			}
		}
	}
}

// At file level any label written where control goes on to the node counts,
// whether a jump names it or not.
func TestReachableAtFileLevel(t *testing.T) {
	src := `<?php namespace N {
	return;
	if ($c) { top: m1(); }
	return;
	if ($c) { other: x(); } else { m2(); }
	return;
	{ ended: return; } m3();
	return;
	{ open: x(); } m4();
	return;
	if ($c) { m5(); late: x(); }
	return;
	switch ($c) { case m6(): return; inner: x(); }
	}`
	// m6: the label of its case is written after it, so that nothing leads
	// into the switch before it.
	want := map[string]bool{"m1": true, "m2": false, "m3": false, "m4": true, "m5": false, "m6": false}
	got := reachableByName(Parse(t, src), syntax.EnclosingVariableScope)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}

// The labels a scope's jumps reach are found whatever the order they are
// written in: a jump waits where a list stops it until a label of that list
// is reached, and one that no label can free is dropped.
func TestReachableThroughJumpsInAnyOrder(t *testing.T) {
	src := `<?php
	function a($c) { if ($c) { goto z; } return 1; a: goto b; b: p1(); return 2; z: goto a; }
	function b($c) { if ($c) { goto z; } return 1; { in: p2(); goto out; return 3; } out: p3(); return 4; z: goto in; }
	function c($c) { if ($c) { return 1; goto a; } return 2; a: p4(); }
	function d($c) { if ($c) { goto z; } return 1; if ($c) { goto q; } late: p5(); return 2; q: p6(); return 3; z: goto late; }
	function e($c) { if ($c) { goto a; } if ($c) { goto a; } return 1; a: p7(); }
	function f($c) { if ($c) { goto z; } return 1; if ($c) { a: x(); } else { while ($c) { p8(); } } return 2; z: goto a; }
	function g($c) { if ($c) { goto z; } return 1; if ($c) { a: x(); if ($c) { x(); } else { p9(); } } return 2; z: goto a; }
	function h($c) { return 1; one: q1(); goto two; two: q2(); goto one; }
	function i($c) { if ($c) { goto z; } return 1; switch ($c) { case 1: return 2; case q3(): a: x(); } return 3; z: goto a; }
	`
	want := map[string]bool{
		"p1": true, "p2": true, "p3": true, "p4": false, "p5": true, "p6": false, "p7": true, "p8": false, "p9": true,
		"q1": false, "q2": false, "q3": false,
	}
	got := reachableByName(Parse(t, src), syntax.EnclosingVariableScope)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
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
		start := testbudget.CPU()
		got := reachableCount(calls)
		elapsed := testbudget.CPU() - start
		if got != want {
			t.Errorf("labelled=%v: %d reachable calls, want %d", labelled, got, want)
		}
		if limit := testbudget.Of(2 * time.Second); elapsed > limit {
			t.Errorf("labelled=%v: %v for 20,000 dead statements (limit %v)", labelled, elapsed, limit)
		}
	}
}

// stateMachine is a function of n labelled states that jump to each other,
// as a generated lexer is written, in an endless loop when looped: every
// state is reached from the first when entered, none when a `return` is
// written before them all.
func stateMachine(n int, entered, looped bool) string {
	var b strings.Builder
	b.WriteString("<?php function f(int $n) { ")
	if !entered {
		b.WriteString("return 0; ")
	}
	if looped {
		b.WriteString("for (;;) { ")
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "s%d: if ($n == 1) { goto fin; } probe($n); if ($n == 2) { goto s%d; } if ($n == 3) { goto s%d; } goto s%d;\n", i, (i*7+3)%n, (i*13+5)%n, (i+1)%n)
	}
	b.WriteString("fin: return $n; }")
	if looped {
		b.WriteString("}")
	}
	return b.String()
}

// labelledTail is a function of n jumps to a label that returns, then n
// labelled statements no jump names.
func labelledTail(n int) string {
	var b strings.Builder
	b.WriteString("<?php function f(int $n) { ")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "if ($n == %d) { goto live; }\n", i)
	}
	b.WriteString("live: return 0;\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "d%d: probe($n);\n", i)
	}
	b.WriteString("}")
	return b.String()
}

// labelledBlocks is a function of depth blocks, one in the other, each
// entered by a jump to a label written after a `return`, with n statements
// in the innermost.
func labelledBlocks(depth, n int) string {
	var b strings.Builder
	b.WriteString("<?php function f(int $n) { ")
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&b, "if ($n == %d) { goto b%d; } return 0; b%d: {\n", i, i, i)
	}
	for i := 0; i < n; i++ {
		b.WriteString("probe($n);\n")
	}
	b.WriteString(strings.Repeat("}", depth+1))
	return b.String()
}

// deadRing is a function of n labelled statements after a `return`, each
// reached only by a jump written in the one before, in a block or not.
func deadRing(n int, block bool) string {
	var b strings.Builder
	b.WriteString("<?php function f(int $n) { return 0; ")
	if block {
		b.WriteString("{ ")
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "r%d: probe($n); if ($n) { goto r%d; }\n", i, (i+1)%n)
	}
	if block {
		b.WriteString("}")
	}
	b.WriteString("}")
	return b.String()
}

// TestReachableManyLabelsAndJumpsLinear guards the landings of a scope:
// asking Reachable of every node of a function with thousands of labels and
// jumps must stay in proportion to their number (each question looked at
// every jump to every label around it: 19 s for a state machine of 1,000
// states, 62 s for 1,000 labelled statements after 1,000 jumps, 40 s for 50
// labelled blocks one in the other).
func TestReachableManyLabelsAndJumpsLinear(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"state machine", stateMachine(4000, true, false), 4000},
		{"state machine nothing enters", stateMachine(4000, false, false), 0},
		{"state machine in an endless loop", stateMachine(4000, true, true), 4000},
		{"state machine in an endless loop nothing enters", stateMachine(4000, false, true), 0},
		{"labels no jump names", labelledTail(4000), 0},
		{"labelled blocks", labelledBlocks(100, 2000), 2000},
		{"ring nothing enters", deadRing(4000, false), 0},
		{"ring nothing enters, in a block", deadRing(4000, true), 0},
	} {
		calls := probesOf(Parse(t, tc.src))
		start := testbudget.CPU()
		got := reachableCount(calls)
		elapsed := testbudget.CPU() - start
		if got != tc.want {
			t.Errorf("%s: %d reachable calls, want %d", tc.name, got, tc.want)
		}
		if limit := testbudget.Of(2 * time.Second); elapsed > limit {
			t.Errorf("%s: %v (limit %v)", tc.name, elapsed, limit)
		}
	}
}

// BenchmarkReachableDeadTail measures Reachable on every node of a long run
// of statements after a `return`, without and with a label that brings them
// back.
func BenchmarkReachableDeadTail(b *testing.B) {
	benchmarkReachable(b, map[string]string{"plain": deadTail(2000, false), "labelled": deadTail(2000, true)})
}

// BenchmarkReachableJumps measures Reachable on every node of functions
// with many labels and jumps: the landings of the scope are worked out anew
// each time.
func BenchmarkReachableJumps(b *testing.B) {
	benchmarkReachable(b, map[string]string{
		"states": stateMachine(1000, true, false), "looped": stateMachine(1000, true, true), "labels": labelledTail(1000), "blocks": labelledBlocks(50, 500), "ring": deadRing(1000, true),
	})
}

func benchmarkReachable(b *testing.B, sources map[string]string) {
	for name, source := range sources {
		b.Run(name, func(b *testing.B) {
			src := []byte(source)
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

func TestCaseConditionsPrecedeTheirBodies(t *testing.T) {
	f := Parse(t, `<?php function f($x){switch($x){case checked_return(): return 1; case checked_break(): break; default: return 2;}}`)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && !Reachable(c, syntax.EnclosingFuncLike(c)) {
			t.Errorf("condition %s is unreachable", astquery.CallLastName(c))
		}
		return true
	})
}

func TestFileRootReachability(t *testing.T) {
	for _, src := range []string{`<?php a(); exit; dead();`, `<?php namespace N; a(); return; dead();`, `<?php namespace N { a(); exit; dead(); }`} {
		f := Parse(t, src)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				want := astquery.CallLastName(c) != "dead"
				if got := Reachable(c, nil); got != want {
					t.Errorf("%s %s reachable=%v", src, astquery.CallLastName(c), got)
				}
			}
			return true
		})
	}
}

func BenchmarkParkedJumps(b *testing.B) {
	for _, size := range []int{400, 800, 1600, 3200} {
		var src strings.Builder
		src.WriteString("<?php function f($c){goto begin;return;")
		for range size {
			src.WriteString("if($c)goto begin;")
		}
		for i := range size {
			fmt.Fprintf(&src, "L%d:;", i)
		}
		src.WriteString("begin:;")
		for i := range size {
			fmt.Fprintf(&src, "if($c)goto L%d;", i)
		}
		src.WriteString("probe();}")
		data := []byte(src.String())
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				f := syntax.Parse("bench.php", data, syntax.Options{})
				syntax.InspectFile(f, func(n syntax.Node) bool {
					if c, ok := n.(*syntax.FuncCall); ok {
						Reachable(c, syntax.EnclosingFuncLike(c))
					}
					return true
				})
			}
		})
	}
}

func TestParkedNestedStatementsStayApart(t *testing.T) {
	f := Parse(t, `<?php function f($c){goto begin;return;if($c){return;if($c)goto dead;one:;two:;three:;}else{live:reached();}begin:if($c)goto one;if($c)goto two;if($c)goto three;if($c)goto live;return;dead:unreached();}`)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			want := astquery.CallLastName(c) == "reached"
			if got := Reachable(c, syntax.EnclosingFuncLike(c)); got != want {
				t.Errorf("%s reachable=%v", astquery.CallLastName(c), got)
			}
		}
		return true
	})
}

func TestNamespaceNamePrecedesBody(t *testing.T) {
	f := Parse(t, `<?php namespace App { return; }`)
	namespace := f.Stmts[0].(*syntax.Namespace)
	if !Reachable(namespace.Name, nil) {
		t.Fatal("namespace name followed its terminating body")
	}
}
