package gdjpegqualityoutsidesupportedrange

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
		{"postfix decrement", "<?php function f($i,$p){ $q = 101; $q--; imagejpeg($i,$p,$q); }", 0},
		{"prefix decrement", "<?php function f($i,$p){ $q = 101; --$q; imagejpeg($i,$p,$q); }", 0},
		{"increment", "<?php function f($i,$p){ $q = -2; $q++; imagejpeg($i,$p,$q); }", 0},
		{"parenthesized", "<?php function f($i,$p){ $q = 101; $q--; imagejpeg($i,$p,($q)); }", 0},
		{"loop decrement first", "<?php function f($i,$p){ $q = 101; do { $q--; imagejpeg($i,$p,$q); } while (rand()); }", 0},
		{"file scope", "<?php $q = 101; $q--; imagejpeg($i,$p,$q);", 0},
		{"earlier argument", "<?php function f($i,$p){ $q = 101; imagejpeg($i,$p[$q--],$q); }", 0},
		{"by-reference closure", "<?php function f($i,$p){ $q = 101; $s = function () use (&$q) { $q--; }; imagejpeg($i,$p,$q); }", 0},
		{"by-reference closure assignment", "<?php function f($i,$p){ $q = 101; $s = function () use (&$q) { $q = 50; }; $s(); imagejpeg($i,$p,$q); }", 0},
		{"copied after decrement", "<?php function f($i,$p){ $b = 101; $b--; $m = $b; $q = $m; imagejpeg($i,$p,$q); }", 0},
		{"extract", "<?php function f($i,$p,$o){ $q = 101; extract($o); imagejpeg($i,$p,$q); }", 0},
		{"parse_str into scope", "<?php function f($i,$p,$o){ $q = 101; parse_str($o); imagejpeg($i,$p,$q); }", 0},
		{"parse_str into result", "<?php function f($i,$p,$o){ $q = 150; parse_str($o, $r); imagejpeg($i,$p,$q); }", 1},
		{"copied plain", "<?php function f($i,$p){ $b = 150; $q = $b; $b = $q; imagejpeg($i,$p,$q); }", 1},
		{"extract after call", "<?php function f($i,$p,$o){ $q = 150; imagejpeg($i,$p,$q); extract($o); }", 1},
		{"compound assignment", "<?php function f($i,$p){ $q = 101; $q -= 5; imagejpeg($i,$p,$q); }", 0},
		{"decrement after call", "<?php function f($i,$p){ $q = 150; imagejpeg($i,$p,$q); $q--; }", 1},
		{"loop decrement last", "<?php function f($i,$p){ $q = 150; while (rand()) { imagejpeg($i,$p,$q); $q--; } }", 0},
		{"other variable", "<?php function f($i,$p){ $q = 150; $o = 1; $o--; $a[0]--; imagejpeg($i,$p,$q); }", 1},
		{"unrelated closure", "<?php function f($i,$p){ $q = 150; $s = function () { $q = 1; $q--; }; imagejpeg($i,$p,$q); }", 1},
		{"by-value closure", "<?php function f($i,$p){ $q = 150; $s = function () use ($q, &$o) { $q--; }; imagejpeg($i,$p,$q); }", 1},
		{"nested function", "<?php $q = 150; function g() { $q = 1; $q--; } imagejpeg($i,$p,$q);", 1},
		{"literal", "<?php function f($i,$p){ $q = 1; $q--; imagejpeg($i,$p,150); }", 1},
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
