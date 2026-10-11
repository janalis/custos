package caseinsensitivestringfunctionsmissuse

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
		{"parameter default", "<?php function f($p, $sep = '/') { return stripos($p, $sep); }", 0, ""},
		{"conditional reassignment", "<?php function f($p, $o, $c) { $n = '/'; if ($c) { $n = $o; } return stripos($p, $n); }", 0, ""},
		{"foreach target", "<?php function f($p, $s) { $n = '-'; foreach ($s as $n) { stripos($p, $n); } }", 0, ""},
		{"named arguments", "<?php function f() { return stripos(needle: 'ABC', haystack: '/'); }", 0, ""},
		{"extract", "<?php function f($p, $o) { $n = '#'; extract($o); return stripos($p, $n); }", 0, ""},
		{"by-reference argument", "<?php function f($p) { $n = '#'; preg_match('/b/', $p, $n); return stripos($p, $n); }", 0, ""},
		{"reference alias", "<?php function f($p) { $n = '#'; $r = &$n; $r = 'b'; return stripos($p, $n); }", 0, ""},
		{"chained variable", "<?php function f($p) { $m = '#'; $n = $m; return stripos($p, $n); }", 0, ""},
		{"private property", "<?php class C { private $s = '-'; function f($p) { return stripos($p, $this->s); } }", 0, ""},
		{"two definitions", "<?php if ($x) { define('DS', 'b'); } else { define('DS', '/'); } function f($p) { return stripos($p, DS); }", 0, ""},
		{
			"single assignment", "<?php function f($p) { $n = '#'; $x = explode($n, $p); return stripos($p, $n); }", 1,
			"<?php function f($p) { $n = '#'; $x = explode($n, $p); return strpos($p, $n); }",
		},
		{"file constant", "<?php const DS = '/'; function f($p) { return stripos($p, DS); }", 1, "<?php const DS = '/'; function f($p) { return strpos($p, DS); }"},
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
