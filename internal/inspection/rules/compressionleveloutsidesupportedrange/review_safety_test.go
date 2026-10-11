package compressionleveloutsidesupportedrange

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
		{"literal", "<?php function f($d) { return gzencode($d, 10); }", 1},
		{"unchanged local", "<?php function f($d) { $l = 10; return gzencode($d, $l); }", 1},
		{"post decrement", "<?php function f($d) { $l = 10; $l--; return gzencode($d, $l); }", 0},
		{"pre decrement", "<?php function f($d) { $l = 10; --$l; return gzencode($d, $l); }", 0},
		{"decrement in branch", "<?php function f($d) { $l = 10; if ($d) { $l--; } return gzencode($d, $l); }", 0},
		{"copy of decremented", "<?php function f($d) { $a = 10; $a--; $b = $a; return gzencode($d, $b); }", 0},
		{"by-reference closure", "<?php function f($d) { $l = 20; $c = function () use (&$l) { $l = 9; }; $c(); return gzencode($d, $l); }", 0},
		{"copy after by-reference closure", "<?php function f($d) { $l = 20; $c = function () use (&$l) { $l = 9; }; $c(); $m = $l; return gzencode($d, $m); }", 0},
		{"by-value closure", "<?php function f($d) { $l = 20; $c = function () use ($l) { return $l; }; return gzencode($d, $l); }", 1},
		{"arrow function copy", "<?php function f($d) { $l = 20; $c = fn() => $l = 3; return gzencode($d, $l); }", 1},
		{"extract", "<?php function f($d) { $l = 20; extract(['l' => 1]); return gzencode($d, $l); }", 0},
		{"shadowed extract", "<?php namespace App; function extract($a) {} function f($d) { $l = 20; extract(['l' => 1]); return \\gzencode($d, $l); }", 1},
		{"dynamic decrement", "<?php function f($d) { $l = 10; $n = 'l'; $$n--; return gzencode($d, $l); }", 0},
		{"global across call", "<?php function f($d) { global $l; $l = 20; g(); return gzencode($d, $l); }", 0},
		{"global without call", "<?php function f($d) { global $l; $l = 20; return gzencode($d, $l); }", 1},
		{"local across call", "<?php function f($d) { $l = 20; g(); return gzencode($d, $l); }", 1},
		{"top level across call", "<?php $l = 20; g(); gzencode('x', $l);", 0},
		{"top level across builtin", "<?php $l = 20; strlen('x'); gzencode('x', $l);", 1},
		{"top level decrement", "<?php $l = 10; $l--; gzencode('x', $l);", 0},
		{"decrement after call", "<?php function f($d) { $l = 10; $r = gzencode($d, $l); $l--; return $r; }", 0},
		{"by-reference closure never invoked", "<?php function f($d) { $l = 20; $c = function () use (&$l) { $l = 9; }; return gzencode($d, $l); }", 0},
		{"extract after call", "<?php function f($d) { $l = 20; $r = gzencode($d, $l); extract(['l' => 1]); return $r; }", 0},
		{"copy of later decremented", "<?php function f($d) { $a = 10; $b = $a; $r = gzencode($d, $b); $a--; return $r; }", 0},
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
