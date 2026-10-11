package jsoninvaliddepth

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
		{"post increment", "<?php function f($j) { $d = 0; $d++; return json_decode($j, true, $d); }", 0},
		{"pre increment", "<?php function f($j) { $d = 0; ++$d; return json_encode($j, 0, $d); }", 0},
		{"parenthesised depth", "<?php function f($j) { $d = 0; $d++; return json_decode($j, true, ($d)); }", 0},
		{"decrement below limit", "<?php function f($j) { $d = 2147483648; $d--; return json_decode($j, true, $d); }", 0},
		{"top level increment", "<?php $d = 0; $d++; json_decode('[]', true, $d);", 0},
		{"while counter", "<?php function f($j) { $d = 0; while ($d < 2) { $d++; json_decode($j, true, $d); } }", 0},
		{"step after call in loop", "<?php function f($j) { $d = 0; do { if ($d) { json_decode($j, true, $d); } $d++; } while ($d < 3); }", 0},
		{"for counter", "<?php function f($j) { for ($d = 0; $d < 3; $d++) { if ($d) { json_decode($j, true, $d); } } }", 0},
		{"by-reference capture", "<?php function f($j) { $d = 0; $g = function () use (&$d) { $d = 4; }; $g(); return json_decode($j, true, $d); }", 0},
		{"foreach target", "<?php function f($j) { $d = 0; foreach ($j as $d) {} return json_decode('[]', true, $d); }", 0},
		{"catch binding", "<?php function f($j) { $d = 0; try { g(); } catch (E $d) {} return json_decode($j, true, $d); }", 0},
		{"extract", "<?php function f($j) { $d = 0; extract($j); return json_decode('[]', true, $d); }", 0},
		{"unchanged", "<?php function f($j) { $d = 0; return json_decode($j, true, $d); }", 1},
		{"unchanged top level", "<?php $d = -1; json_encode([], 0, $d);", 1},
		{"other variable stepped", "<?php function f($j) { $d = 0; $n = 0; $n++; $a[0]++; return json_decode($j, true, $d); }", 1},
		{"stepped after call", "<?php function f($j) { $d = 0; $r = json_decode($j, true, $d); $d++; return $r; }", 1},
		{"stepped in another function", "<?php function g() { $d = 0; $d++; } function f($j) { $d = 0; return json_decode($j, true, $d); }", 1},
		{"value capture", "<?php function f($j) { $d = 0; $g = function () use ($d, &$j) {}; return json_decode($j, true, $d); }", 1},
		{"unrelated foreach and catch", "<?php function f($j) { $d = 0; foreach ($j as $k => $v) {} try { g(); } catch (E $x) {} catch (F) {} return json_decode('[]', true, $d); }", 1},
		{"unchanged in loop", "<?php function f($js) { $d = 0; foreach ($js as $j) { json_decode($j, true, $d); } }", 1},
		{"by-reference capture before any call", "<?php function f($j) { $d = 0; $g = function () use (&$d) { $d = 4; }; return json_decode($j, true, $d); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
