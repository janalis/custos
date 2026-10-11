package staticclosurecanbeused

import (
	"testing"

	"custos/internal/fixing"
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
		fixed        string
	}{
		{"include in body", "<?php class V { function r($f) { return (function () use ($f) { include $f; })(); } }", 0, ""},
		{"include without an object", "<?php spl_autoload_register(function ($c) { require __DIR__ . '/' . $c . '.php'; });", 1, "<?php spl_autoload_register(static function ($c) { require __DIR__ . '/' . $c . '.php'; });"},
		{"include in static method", "<?php class V { static function r($f) { return (function () use ($f) { include $f; })(); } }", 1, "<?php class V { static function r($f) { return (static function () use ($f) { include $f; })(); } }"},
		{"assignment stored on", "<?php class E { function f() { $this->h = $c = function () { return 1; }; } }", 0, ""},
		{"assignment returned", "<?php function f() { return $c = function () { return 1; }; }", 0, ""},
		{"coalesce into property", "<?php class E { function f($h) { $this->h = $h ?? function () { return 1; }; } }", 0, ""},
		{"short ternary returned", "<?php function f($c) { return $c ?: function () { return 1; }; }", 0, ""},
		{"receiver bound", "<?php (function () { return 1; })->call($o);", 0, ""},
		{"receiver unbound", "<?php (function () { return 1; })->bindTo(null);", 1, "<?php (static function () { return 1; })->bindTo(null);"},
		{"file variable bound", "<?php $r = function () { return 1; }; $r->bindTo($o);", 0, ""},
		{"file variable invoked", "<?php $r = function () { return 1; }; $r();", 1, "<?php $r = static function () { return 1; }; $r();"},
		{"file variable, open globals", "<?php $r = function () { return 1; }; extract($vars);", 0, ""},
		{"user function consumer", "<?php function test($l, $b) {} test('a', function () { return 1; });", 1, ""},
		{"unknown static consumer", "<?php Unknown::macro('a', function () { return 1; });", 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
		})
	}
}
