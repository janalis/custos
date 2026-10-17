package gettypemissuse

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
		{"glued return", "<?php function f($x) { return'string'===gettype($x); }", 1, "<?php function f($x) { return is_string($x); }"},
		{"glued negation", "<?php function f($x) { return'string'!==gettype($x); }", 1, "<?php function f($x) { return!is_string($x); }"},
		{
			"glued qualified", "<?php namespace N; use function P\\is_int; function f($x) { return'integer'===gettype($x); }", 1,
			"<?php namespace N; use function P\\is_int; function f($x) { return \\is_int($x); }",
		},
		{"file start", "<?php\n'string'===gettype($x);", 1, "<?php\nis_string($x);"},
		{"by-ref argument", "<?php function p(&$o) {} function f($v) { $t = 'array'; p($t); return gettype($v) === $t; }", 0, ""},
		{"by-value argument", "<?php function p($o) {} function f($v) { $t = 'array'; p($t); return gettype($v) === $t; }", 1, "<?php function p($o) {} function f($v) { $t = 'array'; p($t); return is_array($v); }"},
		{"alias", "<?php function f($v) { $t = 'array'; $r = &$t; $r = 'string'; return gettype($v) === $t; }", 0, ""},
		{"extract", "<?php function f($v, $o) { $t = 'array'; extract($o); return gettype($v) === $t; }", 0, ""},
		{"variable variable", "<?php function f($v, $n) { $t = 'array'; $$n = 'string'; return gettype($v) === $t; }", 0, ""},
		{"invalid name may be overwritten", "<?php function p(&$o) {} function f($v) { $t = 'int'; p($t); return gettype($v) === $t; }", 0, ""},
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
