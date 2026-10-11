package obgetcleancanbeused

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
		{"ternary branch", "<?php $out = $keep ? ob_get_contents() : ''; ob_end_clean();", 0, ""},
		{"ternary condition", "<?php $out = ob_get_contents() ? 1 : 2; ob_end_clean();", 1, ""},
		{"coalesce right", "<?php $out = $a ?? ob_get_contents(); ob_end_clean();", 0, ""},
		{"match arm", "<?php $out = match ($k) { 1 => ob_get_contents(), default => '' }; ob_end_clean();", 0, ""},
		{"match subject", "<?php $out = match (ob_get_contents()) { '' => 0, default => 1 }; ob_end_clean();", 1, ""},
		{"nullsafe argument", "<?php $o?->put(ob_get_contents()); ob_end_clean();", 0, ""},
		{"method argument", "<?php $o->put(ob_get_contents()); ob_end_clean();", 1, ""},
		{"print", "<?php print ob_get_contents(); ob_end_clean();", 1, ""},
		{"later operand", "<?php $r = ob_get_contents() . ob_get_level(); ob_end_clean();", 1, ""},
		{"user call", "<?php log_it(ob_get_contents()); ob_end_clean();", 1, ""},
		{"property target", "<?php $o->p = ob_get_contents(); ob_end_clean();", 1, ""},
		{"compound target", "<?php $s .= ob_get_contents(); ob_end_clean();", 1, ""},
		{"namespaced wrapper", "<?php namespace A; function trim($s) { echo $s; } $s = trim(\\ob_get_contents()); \\ob_end_clean();", 1, ""},
		{"callable read", "<?php $f = ob_get_contents(...); ob_end_clean();", 0, ""},
		{"callable end", "<?php $f = ob_get_contents(); ob_end_clean(...);", 0, ""},
		{"plain", "<?php $s = ob_get_contents(); ob_end_clean();", 1, "<?php $s = ob_get_clean();"},
		{"inert wrappers", "<?php $s = (int) strlen(ob_get_contents()); ob_end_clean();", 1, "<?php $s = (int) strlen(ob_get_clean());"},
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
