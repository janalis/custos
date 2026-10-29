package randomapimigration

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
		findings     int
		fixed        string
	}{
		{"seeded mt_rand", "<?php mt_srand($seed); $a = mt_rand(1, 52);", phpversion.PHP85, 0, ""},
		{"seeded rand", "<?php srand($seed); $a = rand(1, 52);", phpversion.PHP70, 2, "<?php mt_srand($seed); $a = mt_rand(1, 52);"},
		{"reseeded at random", "<?php mt_srand(); $a = mt_rand(1, 52);", phpversion.PHP85, 1, ""},
		{"reversed modern", "<?php $a = rand(10, 1);", phpversion.PHP85, 1, ""},
		{"reversed classic", "<?php $a = rand(10, 1);", phpversion.PHP56, 1, ""},
		{"reversed signed", "<?php $a = rand(-1, -10);", phpversion.PHP56, 1, ""},
		{"ordered classic", "<?php $a = rand(1, 10);", phpversion.PHP56, 1, "<?php $a = mt_rand(1, 10);"},
		{"two spreads", "<?php $a = mt_rand(...$none, ...$none);", phpversion.PHP85, 0, ""},
		{"spread tail", "<?php $a = mt_rand(1, ...$rest);", phpversion.PHP85, 0, ""},
		{"rand spreads", "<?php $a = rand(...$none, ...$none);", phpversion.PHP70, 1, "<?php $a = mt_rand(...$none, ...$none);"},
		{"no arguments", "<?php $a = rand();", phpversion.PHP70, 1, "<?php $a = mt_rand();"},
		{"alias", "<?php srand(7); $a = rand(); $b = getrandmax();", phpversion.PHP71, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
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
