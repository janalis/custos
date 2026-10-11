package stringcasemanipulation

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
		fixed        string
	}{
		{"mb upper changes length", "<?php mb_strpos(mb_strtoupper($h), mb_strtoupper($n));", phpversion.PHP85, ""},
		{"mb lower changes length", "<?php mb_strrpos(mb_strtolower($h), mb_strtolower($n));", phpversion.PHP85, ""},
		{"mb literal needle", "<?php mb_strpos(mb_strtolower($h), 'ss');", phpversion.PHP85, ""},
		{"mb conversion, byte search", "<?php strpos(mb_strtolower($h), mb_strtolower($n));", phpversion.PHP85, ""},
		{"byte conversion, mb search", "<?php mb_strpos(strtolower($h), strtolower($n));", phpversion.PHP85, ""},
		{"byte family", "<?php strpos(strtolower($h), strtolower($n));", phpversion.PHP85, "<?php stripos($h, $n);"},
		{"untyped needle on 8.0", "<?php function f($h, $n) { return strpos(strtolower($h), strtolower($n)); }", phpversion.PHP80, "<?php function f($h, $n) { return stripos($h, $n); }"},
		{"untyped needle on 7.4", "<?php function f($h, $n) { return strpos(strtolower($h), strtolower($n)); }", phpversion.PHP74, ""},
		{"int needle on 7.4", "<?php function f(string $h, int $n) { return strrpos(strtoupper($h), strtoupper($n)); }", phpversion.PHP74, ""},
		{"nullable needle on 7.4", "<?php function f(string $h, ?string $n) { return strpos(strtolower($h), strtolower($n)); }", phpversion.PHP74, ""},
		{"string needle on 7.4", "<?php function f($h, string $n) { return strpos(strtolower($h), strtolower($n)); }", phpversion.PHP74, "<?php function f($h, string $n) { return stripos($h, $n); }"},
		{"literal needle on 7.4", "<?php function f($h) { return strpos(strtolower($h), 'a-1'); }", phpversion.PHP74, "<?php function f($h) { return stripos($h, 'a-1'); }"},
		{"untyped kept needle on 7.4", "<?php function f($n) { return strpos('ABC', strtoupper($n)); }", phpversion.PHP74, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte(tc.source)
			opt := syntax.Options{Version: tc.php}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != 1 {
				t.Fatalf("findings: got %v, want 1", got)
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
