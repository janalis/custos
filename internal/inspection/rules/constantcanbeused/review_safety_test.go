package constantcanbeused

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
		// A suffixed PHP_VERSION ("8.1.2-1ubuntu2.14") sorts above '8.1.2'.
		{"full gt", "<?php version_compare(PHP_VERSION, '8.1.2', '>');", 1, ""},
		{"full le", "<?php version_compare(PHP_VERSION, '8.1.2', '<=');", 1, ""},
		{"full eq", "<?php version_compare(PHP_VERSION, '8.1.2', '==');", 1, ""},
		{"full ne", "<?php version_compare(PHP_VERSION, '8.1.2', '!=');", 1, ""},
		{"full lt", "<?php version_compare(PHP_VERSION, '8.1.2', '<');", 1, "<?php PHP_VERSION_ID < 80102;"},
		{"full ge", "<?php version_compare(PHP_VERSION, '8.1.2', '>=');", 1, "<?php PHP_VERSION_ID >= 80102;"},
		{"short gt", "<?php version_compare(PHP_VERSION, '8.1', '>');", 1, "<?php PHP_VERSION_ID >= 80100;"},
		{"short le", "<?php version_compare(PHP_VERSION, '8', '<=');", 1, "<?php PHP_VERSION_ID < 80000;"},
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
