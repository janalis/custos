package gdimagedimensionsnotpositive

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
		{"post increment", "<?php function f() { $w = 0; $w++; return imagecreate($w, 5); }", 0},
		{"pre increment", "<?php function f() { $w = 0; ++$w; return imagecreate($w, 5); }", 0},
		{"decrement", "<?php function f() { $w = 0; $w--; return imagecreate($w, 5); }", 0},
		{"foreach counter", "<?php function f(array $l) { $r = 0; foreach ($l as $x) { $r++; } return imagecreatetruecolor(9, $r); }", 0},
		{"while counter", "<?php function f($s) { $r = 0; while ($s->next()) { $r++; } return imagecreatetruecolor(9, $r); }", 0},
		{"for counter", "<?php function f() { $n = 0; for ($k = 0; $k < 3; $k++) { $n++; } return imagecreatetruecolor(9, $n); }", 0},
		{"guarded loop counter", "<?php function f() { $n = 0; while (g()) { if ($n) { return imagecreate($n, 5); } $n++; } }", 0},
		{"copied counter", "<?php function f() { $n = 0; $n++; $m = $n; return imagecreate($m, 5); }", 0},
		{"reference capture", "<?php function f() { $n = 0; $g = function () use (&$n) { $n = 10; }; $g(); return imagecreate($n, 5); }", 0},
		{"catch binding", "<?php function f() { $n = 0; try { g(); } catch (\\Exception $n) { } return imagecreate($n, 5); }", 0},
		{"extract", "<?php function f(array $d) { $n = 0; extract($d); return imagecreate($n, 5); }", 0},
		{"parse_str", "<?php function f($q) { $n = 0; parse_str($q); return imagecreate($n, 5); }", 0},
		{"dynamic increment", "<?php function f($x) { $n = 0; $$x++; return imagecreate($n, 5); }", 0},
		{"dynamic compound", "<?php function f($x) { $n = 0; $$x += 3; return imagecreate($n, 5); }", 0},
		{"global with call", "<?php function f() { global $n; $n = 0; g(); return imagecreate($n, 5); }", 0},
		{"top level with call", "<?php $n = 0; g(); imagecreate($n, 5);", 0},
		{"both dimensions", "<?php function f() { $w = 0; $w++; $h = 0; return imagecreate($w, $h); }", 1},
		{"unrelated counter", "<?php function f() { $n = 0; for ($i = 0; $i < 3; $i++) { } return imagecreate($n, 5); }", 1},
		{"value capture", "<?php function f() { $n = 0; $g = function () use ($n) { $n++; }; return imagecreate($n, 5); }", 1},
		{"arrow local write", "<?php function f() { $n = 0; $g = fn() => $n++; return imagecreate($n, 5); }", 1},
		{"local with call", "<?php function f() { $n = 0; g(); return imagecreate($n, 5); }", 1},
		{"top level without call", "<?php $n = 0; imagecreate($n, 5);", 1},
		{"named arguments", "<?php function f() { $h = -1; return imagecreate(height: $h, width: 5); }", 1},
		{"literal", "<?php imagecreatetruecolor(0, 80);", 1},
		{"increment after call", "<?php function f() { $n = 0; $r = imagecreate($n, 5); $n++; return $r; }", 0},
		{"reference capture never invoked", "<?php function f() { $n = 0; $g = function () use (&$n) { $n = 10; }; return imagecreate($n, 5); }", 0},
		{"extract after call", "<?php function f(array $d) { $n = 0; $r = imagecreate($n, 5); extract($d); return $r; }", 0},
		{"copy of later incremented", "<?php function f() { $n = 0; $m = $n; $r = imagecreate($m, 5); $n++; return $r; }", 0},
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
