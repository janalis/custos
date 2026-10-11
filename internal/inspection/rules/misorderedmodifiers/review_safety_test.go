package misorderedmodifiers

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
		{"keyword in comment, ordered", "<?php class A { public /* not final on purpose */ static function f() {} }", 0, ""},
		{"visibility in comment, ordered", "<?php class A { final /* private */ public function f() {} }", 0, ""},
		{"visibility in comment", "<?php class A { static /* was protected */ public function f() {} }", 1, "<?php class A { public /* was protected */ static function f() {} }"},
		{"line comment", "<?php class A { static // final\n public function f() {} }", 1, "<?php class A { public // final\n static function f() {} }"},
		{"no comment collapses", "<?php class A { static\n public function f() {} }", 1, "<?php class A { public static function f() {} }"},
		{"single modifier", "<?php class A { static function f() {} }", 0, ""},
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
