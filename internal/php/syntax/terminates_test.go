package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestTerminates(t *testing.T) {
	cases := map[string]bool{
		`return;`:                      true,
		`throw new E;`:                 true,
		`(exit());`:                    true,
		`f();`:                         false,
		`{ f(); return; }`:             true,
		`if ($a) return; else return;`: true,
		`if ($a) return; elseif ($b) f(); else return;`:    false,
		`if ($a) return; elseif ($b) return; else return;`: true,
		`if ($a) return;`:                                                false,
		`try { f(); } finally { return; }`:                               true,
		`try { return; } catch (E $e) { return; }`:                       true,
		`try { return; } catch (E $e) { f(); }`:                          false,
		`try { f(); } catch (E $e) { return; }`:                          false,
		`switch ($a) { case 1: return; default: return; }`:               true,
		`switch ($a) { case 1: return; }`:                                false,
		`switch ($a) { case 1: break; default: return; }`:                false,
		`switch ($a) { default: f(); }`:                                  false,
		`while (true) { f(); }`:                                          true,
		`while (TRUE) { if ($a) break; }`:                                false,
		`while ($a) { }`:                                                 false,
		`do { } while (true);`:                                           true,
		`do { return; } while ($a);`:                                     true,
		`do { break; } while ($a);`:                                      false,
		`do { continue; } while (false);`:                                false,
		`do { if ($q) { break; } return 0; } while ($a);`:                false,
		`do { if ($q) { continue; } return 0; } while ($a);`:             false,
		`do { if ($q) { continue $n; } return 0; } while ($a);`:          false,
		`do { continue; } while (true);`:                                 true,
		`do { goto out; } while ($a);`:                                   false,
		`do { goto top; top: return; } while ($a);`:                      true,
		`do { foreach ($l as $x) { continue; } return; } while ($a);`:    true,
		`do { switch ($q) { case 1: continue 2; } return; } while ($a);`: false,
		`for (;;) { }`:                 true,
		`for (;$a;) { }`:               false,
		`declare(ticks=1) { return; }`: true,
		`declare(ticks=1);`:            false,
		// breaks nested in inner loops/switches and function-likes
		`while (true) { foreach ($a as $x) { break; } }`:                          true,
		`while (true) { foreach ($a as $x) { break 2; } }`:                        false,
		`while (true) { foreach ($a as $x) { break 0x2; } }`:                      false,
		`while (true) { foreach ($a as $x) { break 01; } }`:                       true,
		`while (true) { foreach ($a as $x) { break 0; } }`:                        true,
		`while (true) { foreach ($a as $x) { foreach ($b as $y) { break 3; } } }`: false,
		`while (true) { foreach ($a as $x) { foreach ($b as $y) { break 2; } } }`: true,
		`while (true) { foreach ($a as $x) { break 99999999999; } }`:              false,
		`while (true) { foreach ($a as $x) { break $n; } }`:                       false,
		`while (true) { $f = function () { while (1) break; }; $g = fn() => 1; }`: true,
		`while (true) { continue; }`:                                              true,
		// continue targeting a switch acts as break
		`switch ($a) { case 1: continue; default: return; }`:                          false,
		`switch ($a) { case 1: foreach ($b as $c) { continue; } default: return; }`:   true,
		`switch ($a) { case 1: foreach ($b as $c) { continue 2; } default: return; }`: false,
		`switch ($a) { case 1: continue 2; default: return; }`:                        true,
	}
	for src, want := range cases {
		f := parse(t, "<?php "+src, phpversion.PHP84)
		if len(f.Errors) > 0 {
			t.Fatalf("%s: %v", src, f.Errors)
		}
		if got := Terminates(f.Stmts[0]); got != want {
			t.Errorf("%s: Terminates = %v, want %v", src, got, want)
		}
	}
	if Terminates(nil) || containsBreak(nil) {
		t.Fatal("nil")
	}
}

// An endless loop is one that control never falls out of wherever it entered
// it: a loop that only ends because its body always returns is not one.
func TestEndless(t *testing.T) {
	cases := map[string]bool{
		`while (true) { f(); }`:                   true,
		`while (true) { if ($a) { break; } }`:     false,
		`while ($a) { return; }`:                  false,
		`for (;;) { f(); }`:                       true,
		`for (;;) { if ($a) { break; } }`:         false,
		`for (; $a;) { return; }`:                 false,
		`do { f(); } while (true);`:               true,
		`do { if ($a) { break; } } while (true);`: false,
		`do { return; } while ($a);`:              false,
		`foreach ($a as $b) { return; }`:          false,
		`{ return; }`:                             false,
	}
	for src, want := range cases {
		f := parse(t, "<?php "+src, phpversion.PHP84)
		if len(f.Errors) > 0 {
			t.Fatalf("%s: %v", src, f.Errors)
		}
		for range 2 {
			// The second answer is the one kept on the loop.
			if got := Endless(f.Stmts[0]); got != want {
				t.Errorf("%s: Endless = %v, want %v", src, got, want)
			}
		}
	}
}

func TestStmtIndexAndFirstTerminating(t *testing.T) {
	f := parse(t, "<?php f(); return; g();", phpversion.PHP84)
	if StmtIndex(f.Stmts, f.Stmts[2]) != 2 || StmtIndex(nil, f.Stmts[0]) != -1 || StmtIndex(f.Stmts, nil) != -1 {
		t.Fatal("StmtIndex basics")
	}
	// A non-statement sharing a statement's start is not in the list.
	call := f.Stmts[0].(*ExprStmt).Expr
	if StmtIndex(f.Stmts, call) != -1 {
		t.Fatal("expression found as statement")
	}
	// Out-of-order list (not produced by the parser): linear fallback.
	rev := []Stmt{f.Stmts[2], f.Stmts[1], f.Stmts[0]}
	if StmtIndex(rev, f.Stmts[0]) != 2 {
		t.Fatal("linear fallback")
	}
	other := parse(t, "<?php h();", phpversion.PHP84)
	if StmtIndex(f.Stmts, other.Stmts[0]) != -1 {
		t.Fatal("foreign statement")
	}
	ns := parse(t, "<?php namespace A; f(); return;", phpversion.PHP84)
	first := FirstTerminating(ns.Stmts[0])
	if first != 1 || FirstTerminating(ns.Stmts[0]) != first { // second call: memoised
		t.Fatal("namespace list")
	}
	sw := parse(t, "<?php switch ($a) { case 1: f(); }", phpversion.PHP84)
	if FirstTerminating(sw.Stmts[0].(*Switch).Cases[0]) != 1 || FirstTerminating(sw.Stmts[0]) != 0 {
		t.Fatal("case list / other node")
	}
	if l, ok := StmtListOf(sw.Stmts[0].(*Switch).Cases[0]); !ok || len(l) != 1 {
		t.Fatal("StmtListOf case")
	}
	if _, ok := StmtListOf(ns.Stmts[0]); !ok {
		t.Fatal("StmtListOf namespace")
	}
	if _, ok := StmtListOf(sw.Stmts[0]); ok {
		t.Fatal("StmtListOf switch")
	}
}

func TestEndlessLiteralTruthAndGotos(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{
		{`for (; 1;) {}`, true},
		{`for (; f(), "yes";) {}`, true},
		{`for (; true, 0;) {}`, false},
		{`while (1) {}`, true},
		{`while (0x10) {}`, true},
		{`while (0b10) {}`, true},
		{`while (0o10) {}`, true},
		{`while (1_000) {}`, true},
		{`while (0) {}`, false},
		{`while ('') {}`, false},
		{`while ("0") {}`, false},
		{`while ("\x30") {}`, false},
		{`while ("\060") {}`, false},
		{`while (b"") {}`, false},
		{`while (B'0') {}`, false},
		{`while ('\0') {}`, true},
		{`while ('0.0') {}`, true},
		{`while ("\n") {}`, true},
		{`while (!0) {}`, true},
		{`while (!false) {}`, true},
		{`while (!true) {}`, false},
		{`while (!null) {}`, true},
		{`while (!f()) {}`, false},
		{`while (-1) {}`, true},
		{`while (+2) {}`, true},
		{`while (-0) {}`, false},
		{`while (-$x) {}`, false},
		{`while (1.5) {}`, false},
		{`while (OTHER) {}`, false},
		{`while (true) { if ($x) { goto outside; } } outside: done();`, false},
		{`while (true) { goto missing; }`, false},
		{`while (true) { goto inside; inside: next(); }`, true},
		{`while (true) { inside: next(); goto inside; }`, true},
		{`while (true) { while ($x) { goto outside; } } outside: done();`, false},
		{`while (true) { function nested() { goto outside; } } outside: done();`, true},
		{`while (true) { $f = function () { goto outside; }; } outside: done();`, true},
		{`while (true) { if ($x) { return; } }`, true},
		{`while (true) { if ($x) { throw new E; } }`, true},
		{`do { goto outside; } while (true); outside: done();`, false},
		{`for (;;) { goto outside; } outside: done();`, false},
	} {
		f := parse(t, "<?php "+tc.source, phpversion.PHP85)
		if len(f.Errors) != 0 {
			t.Fatalf("%s: %v", tc.source, f.Errors)
		}
		for range 2 {
			if got := Endless(f.Stmts[0]); got != tc.want {
				t.Errorf("%s: %v want %v", tc.source, got, tc.want)
			}
		}
	}
	if isTrueConst(&ConstFetch{}) {
		t.Fatal("nameless constant")
	}
	if truth, known := constantTruth(&Literal{Raw: "not an int", LitKind: LitInt}); truth || known {
		t.Fatal("invalid integer")
	}
}

func BenchmarkEndlessLiteralAndGoto(b *testing.B) {
	for _, source := range []string{`while (1) { next(); }`, `while ("\x31") { next(); }`, `while (true) { goto inside; inside: next(); }`} {
		b.Run(source, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				f := Parse("bench.php", []byte("<?php "+source), Options{Version: phpversion.PHP85})
				Endless(f.Stmts[0])
			}
		})
	}
}
