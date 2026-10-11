package substrusedasstrpos

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
		{"untyped strict", "<?php function f($s, $p) { return substr($s, 0, strlen($p)) === $p; }", phpversion.PHP85, ""},
		{"typed strict", "<?php function f($s, string $p) { return substr($s, 0, strlen($p)) === $p; }", phpversion.PHP85, "<?php function f($s, string $p) { return strpos($s, $p) === 0; }"},
		{"documented only", "<?php /** @param string $p */ function f($s, $p) { return substr($s, 0, strlen($p)) === $p; }", phpversion.PHP85, ""},
		{"nullable", "<?php function f($s, ?string $p) { return substr($s, 0, strlen($p)) !== $p; }", phpversion.PHP85, ""},
		{"empty needle 7.4", "<?php function f($s, string $p) { return substr($s, 0, strlen($p)) === $p; }", phpversion.PHP74, ""},
		{"empty literal 7.4", "<?php $r = substr($s, 0, 0) === '';", phpversion.PHP74, ""},
		{"empty literal 8.0", "<?php $r = substr($s, 0, 0) === '';", phpversion.PHP80, "<?php $r = strpos($s, '') === 0;"},
		{"literal 7.4", "<?php $r = substr($s, 0, 2) === 'ab';", phpversion.PHP74, "<?php $r = strpos($s, 'ab') === 0;"},
		{"effect", "<?php $r = substr($s, 0, strlen(f())) === f();", phpversion.PHP85, ""},
		{"hex 5.6", "<?php $r = substr($s, 0, 4) == '0x1A';", phpversion.PHP56, ""},
		{"hex 7.0", "<?php $r = substr($s, 0, 4) == '0x1A';", phpversion.PHP70, "<?php $r = strpos($s, '0x1A') === 0;"},
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
		})
	}
}
