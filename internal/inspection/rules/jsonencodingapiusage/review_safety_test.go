package jsonencodingapiusage

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
		{"spread only", "<?php json_decode(...$args);", 0, ""},
		{"spread rest", "<?php json_decode($j, ...$rest);", 0, ""},
		{"spread encode", "<?php json_encode(['a' => 1], ...$o);", 0, ""},
		{"first-class callable", "<?php $f = json_decode(...);", 0, ""},
		{"array flag", "<?php json_decode($j, flags: JSON_OBJECT_AS_ARRAY | JSON_THROW_ON_ERROR);", 0, ""},
		{"array flag alone", "<?php json_decode($j, flags: JSON_OBJECT_AS_ARRAY);", 1, "<?php json_decode($j, flags: JSON_THROW_ON_ERROR | JSON_OBJECT_AS_ARRAY);"},
		{"unknown flags", "<?php json_decode($j, flags: $m | JSON_THROW_ON_ERROR);", 1, ""},
		{
			"known flags", "<?php json_decode($j, flags: JSON_THROW_ON_ERROR);", 1,
			"<?php json_decode($j, flags: JSON_THROW_ON_ERROR, associative: false);",
		},
		{"missing subject", "<?php json_encode(depth: 3);", 0, ""},
		{"positional", "<?php json_decode($j);", 2, "<?php json_decode($j, false);"},
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
