package forkchildfallsintoparentwork

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const fork = "<?php function f($w, $n) { foreach ($w as $x) { $p = pcntl_fork(); if ($p === 0) { "
	for _, tc := range []struct {
		name, child string
		findings    int
	}{
		{"plain fallthrough", "work();echo 1;", 1},
		{"finally exits", "try { work(); } finally { exit(0); }", 0},
		{"every branch exits", "if ($n) { exit(0); } else { throw new Exception(); }", 0},
		{"one branch falls through", "if ($n) { exit(0); }", 1},
		{"nested block", "{ return; }", 0},
		{"trailing empty statement", "exit(0);;", 0},
		{"endless loop", "for (;;) { sleep(1); }", 0},
		{"loop left by break", "while (true) { break; }", 1},
		{"switch exits", "switch ($n) { default: exit(0); }", 0},
		{"continue reaches the parent loop", "work(); continue;", 1},
		{"break reaches past the parent loop", "work(); break;", 1},
		{"inner break stays inside", "while ($n) { break; } exit;", 0},
		{"unknown function", "runAndExit($w);", 0},
		{"unknown method", "$w->runForever();", 0},
		{"parenthesised unknown call", "(runAndExit($w));", 0},
		{"builtin returns", "usleep(10);", 1},
		{"pcntl_exec can fail", "pcntl_exec('/bin/ls');", 1},
		{"new object", "new Worker();", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(fork+tc.child+" } echo 'parent'; } }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			for _, g := range got {
				if len(g.Fixes) != 0 {
					t.Fatalf("unexpected fix: %+v", g)
				}
			}
		})
	}
}

// Flow analysis currently withholds its facts in functions using goto or
// levelled jumps, so those shapes are checked on the helper directly.
func TestJumpsOut(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"goto", "goto out;", true},
		{"levelled break", "while ($n) { break 2; } exit;", true},
		{"levelled continue", "while ($n) { continue 2; } exit;", true},
		{"plain break", "break;", true},
		{"plain continue", "continue;", true},
		{"inner break", "foreach ($w as $x) { break; } exit;", false},
		{"inner continue", "do { continue; } while ($n); exit;", false},
		{"switch break", "switch ($n) { case 1: break; } exit;", false},
		{"closure", "$f = function () { goto out; out: }; exit;", false},
		{"exit", "exit;", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php while ($w) { if ($p === 0) { "+tc.body+" } } out:"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			body := f.Stmts[0].(*syntax.While).Body.(*syntax.Block).Stmts[0].(*syntax.If).Body
			if got := jumpsOut(body, false); got != tc.want {
				t.Fatalf("jumpsOut: got %v, want %v", got, tc.want)
			}
		})
	}
}
