package strstrusedasstrpos

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
		{"not under comparison", "<?php $r = !strstr($h, $n) == $f;", 1, "<?php $r = (strpos($h, $n) === false) == $f;"},
		{"not under cast", "<?php $r = (int) !strstr($h, $n);", 1, "<?php $r = (int) (strpos($h, $n) === false);"},
		{"double not", "<?php $r = !!strstr($h, $n);", 1, "<?php $r = !(strpos($h, $n) === false);"},
		{"not under concat", "<?php $r = !strstr($h, $n) . 'x';", 1, "<?php $r = (strpos($h, $n) === false) . 'x';"},
		{"not under clone", "<?php $r = clone !strstr($h, $n);", 1, "<?php $r = clone (strpos($h, $n) === false);"},
		{"not as condition", "<?php if (!strstr($h, $n)) {}", 1, "<?php if (strpos($h, $n) === false) {}"},
		{"not under coalesce", "<?php $r = $f ?? !strstr($h, $n);", 1, "<?php $r = $f ?? strpos($h, $n) === false;"},
		{"named arguments", "<?php $r = strstr(needle: $n, haystack: $h) ? 1 : 2;", 0, ""},
		{"unpacked argument", "<?php $r = strstr($h, ...$rest) ? 1 : 2;", 0, ""},
		{"first-class callable", "<?php $r = strstr(...) ? 1 : 2;", 0, ""},
		{"prefix as condition", "<?php $r = strstr($h, $n, true) ? 1 : 2;", 0, ""},
		{"prefix loosely compared", "<?php $r = strstr($h, $n, true) != false;", 0, ""},
		{"prefix strictly compared", "<?php $r = strstr($h, $n, true) !== false;", 1, "<?php $r = strpos($h, $n) !== false;"},
		{"explicit default", "<?php $r = strstr($h, $n, false) ? 1 : 2;", 1, "<?php $r = strpos($h, $n) !== false ? 1 : 2;"},
		{"dynamic before_needle", "<?php $r = strstr($h, $n, mode()) === false;", 0, ""},
		{"too many arguments", "<?php $r = strstr($h, $n, false, 1) ? 1 : 2;", 0, ""},
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
