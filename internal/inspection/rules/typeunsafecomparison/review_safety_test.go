package typeunsafecomparison

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		php          phpversion.Version
		fixed        string
	}{
		{"loop back-edge", "<?php function f(array $xs) { $s = 'a'; foreach ($xs as $x) { if ($s == 'yes') { return 1; } $s = $x; } return 0; }", phpversion.PHP85, ""},
		{"reference alias", "<?php function f() { $s = 'a'; $r = &$s; $r = true; return $s == 'yes'; }", phpversion.PHP85, ""},
		{"by-reference builtin", "<?php function f() { $s = 'a'; settype($s, 'bool'); return $s == 'yes'; }", phpversion.PHP85, ""},
		{"by-reference variadic", "<?php function f(string $in) { $m = ''; sscanf($in, '%d', $m); return $m == 'yes'; }", phpversion.PHP85, ""},
		{"global", "<?php function f() { global $g; $g = 'a'; chg(); return $g == 'yes'; }", phpversion.PHP85, ""},
		{"call-time reference", "<?php function f() { $s = 'a'; strlen(&$s); return $s == 'yes'; }", phpversion.PHP53, ""},
		{"top-level code", "<?php $s = 'a'; chg(); $s == 'yes';", phpversion.PHP85, ""},
		{"phpdoc parameter", "<?php /** @param string $flag */ function f($flag) { return $flag == 'yes'; }", phpversion.PHP85, ""},
		{"phpdoc property", "<?php class S { /** @var string */ public $mode = 'off'; function on() { return $this->mode == 'on'; } }", phpversion.PHP85, ""},
		{"phpdoc return", "<?php /** @return string */ function g() { return 'a'; } function f() { return g() == 'yes'; }", phpversion.PHP85, ""},
		{"inline var doc", "<?php function f($x) { /** @var string $s */ $s = $x; return $s == 'yes'; }", phpversion.PHP85, ""},
		{"infinite float", "<?php function f(float $x) { return $x == 'INF'; }", phpversion.PHP85, ""},
		{"negative infinite float", "<?php function f(?float $x) { return '-INF' != $x; }", phpversion.PHP85, ""},
		{"infinite float through a local", "<?php function f($x) { $y = (float) $x; return $y == 'INF'; }", phpversion.PHP85, ""},
		{
			"float against other text", "<?php function f(float $x) { return $x == 'inf'; }", phpversion.PHP85,
			"<?php function f(float $x) { return $x === 'inf'; }",
		},
		{
			"int against INF", "<?php function f(int $x) { return $x == 'INF'; }", phpversion.PHP85,
			"<?php function f(int $x) { return $x === 'INF'; }",
		},
		{
			"native parameter", "<?php function f(string $flag) { return $flag == 'yes'; }", phpversion.PHP85,
			"<?php function f(string $flag) { return $flag === 'yes'; }",
		},
		{
			"native property", "<?php class S { public string $mode = 'off'; function on() { return $this->mode != 'on'; } }", phpversion.PHP85,
			"<?php class S { public string $mode = 'off'; function on() { return $this->mode !== 'on'; } }",
		},
		{
			"stable local", "<?php function f($x) { $s = strtolower($x); if ($x) { $s = 'b'; } return $s == 'yes'; }", phpversion.PHP85,
			"<?php function f($x) { $s = strtolower($x); if ($x) { $s = 'b'; } return $s === 'yes'; }",
		},
		{
			"by-value argument", "<?php function f(string $s) { strlen($s); return $s == 'yes'; }", phpversion.PHP74,
			"<?php function f(string $s) { strlen($s); return $s === 'yes'; }",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte(tc.source)
			opt := syntax.Options{Version: tc.php}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != 1 {
				t.Fatalf("findings: got %v, want 1", got)
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
