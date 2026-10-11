package unsetconstructscanbemerged

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
		{"close tag after each", "<?php unset($a) ?>\n<?php unset($b) ?>\n<p>x</p>", 1, ""},
		{"close tag after last", "<?php unset($a); unset($b) ?>\n<p>x</p>", 1, ""},
		{"close tag after first", "<?php unset($a) ?>\n<?php unset($b);", 1, ""},
		{"close tag later in run", "<?php unset($a); unset($b); unset($c) ?>\n<p>x</p>", 2, ""},
		{"glued tags", "<?php unset($m)?><?php unset($n) ?>\n", 1, ""},
		{"tags between terminated statements", "<?php unset($a); ?>\n<?php unset($b); ?>\n", 1, "<?php unset($a, $b); ?>\n<?php  ?>\n"},
		{"comment in first", "<?php unset($a /* keep */); unset($b);", 1, ""},
		{"comment in merged", "<?php unset($a); unset($b, // keep\n$c);", 1, ""},
		{"comment later in run", "<?php unset($a); unset($b); unset(/* keep */ $c);", 2, ""},
		{"comment between statements", "<?php unset($a); /* why */ unset($b);", 1, "<?php unset($a, $b); /* why */"},
		{"trailing comma", "<?php unset($a,); unset($b,);", 1, "<?php unset($a, $b);"},
		{"run of three", "<?php unset($a); unset($b); unset($c);", 2, "<?php unset($a, $b, $c);"},
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
				t.Fatalf("fixed source: got %q, want %q", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
