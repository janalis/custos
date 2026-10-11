package unnecessaryassertion

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
	const head = "<?php class T { function t($d, $rows) { "
	for _, tc := range []struct {
		name, body string
		findings   int
		fixed      string
	}{
		{"instance receiver", "$d->expects($this->any())->method('x');", 1, "$d->method('x');"},
		{"nullsafe chain kept", "$d?->owner->expects($this->any())->method('x');", 1, "$d?->owner->method('x');"},
		{"static receiver", "static::expects(self::any())->method('x');", 1, ""},
		{"named class receiver", "T::expects(self::any())->method('x');", 1, ""},
		{"variable class receiver", "$d::expects(self::any())->method('x');", 1, ""},
		{"nullsafe receiver", "$d?->expects($this->any())->method('x');", 1, ""},
		{"other matcher", "$d->expects($this->once())->method('x');", 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(head + tc.body + " } }")
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
			want := string(src)
			if tc.fixed != "" {
				want = head + tc.fixed + " } }"
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
