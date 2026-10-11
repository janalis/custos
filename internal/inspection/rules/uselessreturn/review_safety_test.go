package uselessreturn

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
		{"foreach by reference", "<?php function f($xs) { foreach ($xs as &$v) { return $v = 9; } }", 0, ""},
		{"keyed foreach by reference", "<?php function f($xs) { foreach ($xs as $k => &$v) { return $v = 9; } }", 0, ""},
		{"foreach by value", "<?php function f($xs) { foreach ($xs as $v) { return $v = 9; } }", 1, "<?php function f($xs) { foreach ($xs as $v) { return 9; } }"},
		{"reference target", "<?php function f($o) { $v = &$o->p; return $v = 5; }", 0, ""},
		{"reference source", "<?php function f($o) { $o->p = &$v; return $v = 5; }", 0, ""},
		{"other reference", "<?php function f($o) { $w = &$o->p; return $v = 5; }", 1, "<?php function f($o) { $w = &$o->p; return 5; }"},
		{"reference destructuring", "<?php function f($p) { [&$v] = $p; return $v = 5; }", 0, ""},
		{"keyed reference list", "<?php function f($p) { list('k' => &$v) = $p; return $v = 5; }", 0, ""},
		{"reference array item", "<?php function f($o) { $o->r = [&$v]; return $v = 5; }", 0, ""},
		{"by-value array item", "<?php function f($o) { $o->r = [$v]; return $v = 5; }", 1, "<?php function f($o) { $o->r = [$v]; return 5; }"},
		{"reference assignment", "<?php function f(array &$s) { return $c = &$s['k']; }", 0, ""},
		{"outer finally", "<?php function f() { try { try { return $r = 7; } catch (E $e) {} } finally { echo $r ?? 0; } }", 0, ""},
		{"unrelated finally", "<?php function f() { try { try { return $r = 7; } finally { g(); } } finally { h(); } }", 1, "<?php function f() { try { try { return 7; } finally { g(); } } finally { h(); } }"},
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
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
