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
		`if ($a) return;`:                                  false,
		`try { f(); } finally { return; }`:                 true,
		`try { return; } catch (E $e) { return; }`:         true,
		`try { return; } catch (E $e) { f(); }`:            false,
		`try { f(); } catch (E $e) { return; }`:            false,
		`switch ($a) { case 1: return; default: return; }`: true,
		`switch ($a) { case 1: return; }`:                  false,
		`switch ($a) { case 1: break; default: return; }`:  false,
		`switch ($a) { default: f(); }`:                    false,
		`while (true) { f(); }`:                            true,
		`while (TRUE) { if ($a) break; }`:                  false,
		`while ($a) { }`:                                   false,
		`do { } while (true);`:                             true,
		`do { return; } while ($a);`:                       true,
		`for (;;) { }`:                                     true,
		`for (;$a;) { }`:                                   false,
		`declare(ticks=1) { return; }`:                     true,
		`declare(ticks=1);`:                                false,
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
