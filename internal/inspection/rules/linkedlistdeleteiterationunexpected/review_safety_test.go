package linkedlistdeleteiterationunexpected

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
	const setup = "<?php function f($c, $m) { $l = new SplQueue(); $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"proven delete", "foreach ($l as $x) {}", 1},
		{"parenthesized", "foreach (($l) as $x) {}", 1},
		{"unrelated call between", "$l->push(1); echo count($l); foreach ($l as $x) {}", 1},
		{"keep after loop", "foreach ($l as $x) {} $l->setIteratorMode(0);", 1},
		{"keep later in enclosing loop", "while ($c--) { foreach ($l as $x) {} $l->setIteratorMode(0); }", 0},
		{"conditional keep before delete", "if ($c) { $l->setIteratorMode(0); } $l->setIteratorMode(1); foreach ($l as $x) {}", 1},
		{"dominating keep", "$l->setIteratorMode(0); foreach ($l as $x) {}", 0},
		{"keep in if", "if ($c) { $l->setIteratorMode(0); } foreach ($l as $x) {}", 0},
		{"keep in else", "if ($c) { return; } else { $l->setIteratorMode(0); } foreach ($l as $x) {}", 0},
		{"keep in try", "try { $l->setIteratorMode(0); } catch (Exception $e) {} foreach ($l as $x) {}", 0},
		{"keep in switch", "switch ($c) { case 1: $l->setIteratorMode(0); } foreach ($l as $x) {}", 0},
		{"keep in ternary", "$z = $c ? 1 : $l->setIteratorMode(0); foreach ($l as $x) {}", 0},
		{"keep in short circuit", "$c && $l->setIteratorMode(0); foreach ($l as $x) {}", 0},
		{"keep in for header", "for ($i = 0; $i < 2; $l->setIteratorMode(0), $i++) {} foreach ($l as $x) {}", 0},
		{"nullsafe keep", "$l?->setIteratorMode(0); foreach ($l as $x) {}", 0},
		{"dynamic method name", "$l->$m(0); foreach ($l as $x) {}", 0},
		{"closure keep", "$k = function () use ($l) { $l->setIteratorMode(0); }; $k(); foreach ($l as $x) {}", 0},
		{"closure capture only", "$k = function () use ($l) { reset_mode($l); }; $k(); foreach ($l as $x) {}", 0},
		{"reference capture", "$k = function () use (&$l) {}; foreach ($l as $x) {}", 0},
		{"arrow keep", "$k = fn () => $l->setIteratorMode(0); $k(); foreach ($l as $x) {}", 0},
		{"callable array", "$k = [$l, 'setIteratorMode']; $k(0); foreach ($l as $x) {}", 0},
		{"array element", "$a = [$l]; $a[0]->setIteratorMode(0); foreach ($l as $x) {}", 0},
		{"unknown callee", "reset_mode($l); foreach ($l as $x) {}", 0},
		{"alias", "$o = $l; $o->setIteratorMode(0); foreach ($l as $x) {}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(setup+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
