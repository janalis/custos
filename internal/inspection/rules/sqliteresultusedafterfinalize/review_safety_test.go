package sqliteresultusedafterfinalize

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
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"plain statement", "<?php function f(SQLite3Result $r) { $r->finalize(); $r->fetchArray(); }", 1},
		{"queried result", "<?php function f(SQLite3 $db) { $r = $db->query('select 1'); $r->finalize(); return $r->fetchArray(); }", 1},
		{"top level", "<?php $db = new SQLite3(':memory:'); $r = $db->query('select 1'); $r->finalize(); $r->fetchArray();", 1},
		{"stored outcome", "<?php function f(SQLite3Result $r) { $ok = $r->finalize(); if ($ok) { $r->numColumns(); } }", 1},
		{"compared with true", "<?php function f(SQLite3Result $r) { if (($r->finalize()) === \\true) { $r->numColumns(); } }", 1},
		{"loop body", "<?php function f(SQLite3Result $r, $c) { while ($c) { $r->finalize(); $r->reset(); } }", 1},
		{"reassigned", "<?php function f(SQLite3Result $r, SQLite3Stmt $s) { $r->finalize(); $r = $s->execute(); $r->fetchArray(); }", 0},
		{"switch fallthrough", "<?php function f(SQLite3Result $r, $a) { switch ($a) { case 1: $r->finalize(); case 2: $r->fetchArray(); } }", 0},
		{"loop before use", "<?php function f(SQLite3Result $r, $c) { while ($c) { $r->finalize(); } $r->fetchArray(); }", 0},
		{"by-reference guard", "<?php function f(SQLite3Result &$r) { if ($r->finalize()) { $r->fetchArray(); } }", 0},
		{"user subclass method", "<?php class R { function finalize() {} function fetchArray() {} } function f(R $r) { $r->finalize(); $r->fetchArray(); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
