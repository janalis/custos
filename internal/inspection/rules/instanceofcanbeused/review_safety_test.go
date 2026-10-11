package instanceofcanbeused

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
	const head = "<?php final class Invoice {} function f(Invoice $o, ?Invoice $p, bool $b) { return "
	for _, tc := range []struct {
		name, expr, fixed string
	}{
		{"cast", "(int) is_a($o, 'Invoice')", `(int) ($o instanceof \Invoice)`},
		{"bitwise not", "~is_a($o, 'Invoice')", `~($o instanceof \Invoice)`},
		{"silenced", "@is_a($o, 'Invoice')", `@($o instanceof \Invoice)`},
		{"negation", "!is_a($o, 'Invoice')", `!$o instanceof \Invoice`},
		{"power left", "is_a($o, 'Invoice') ** 2", `($o instanceof \Invoice) ** 2`},
		{"power right", "2 ** is_a($o, 'Invoice')", `2 ** ($o instanceof \Invoice)`},
		{"addition", "1 + is_a($o, 'Invoice')", `1 + $o instanceof \Invoice`},
		{"clone", "clone is_a($o, 'Invoice')", `clone ($o instanceof \Invoice)`},
		{"dim base", "is_a($o, 'Invoice')[0]", `($o instanceof \Invoice)[0]`},
		{"dim index", "$b[is_a($o, 'Invoice')]", `$b[$o instanceof \Invoice]`},
		{"member", "is_a($o, 'Invoice')->x", `($o instanceof \Invoice)->x`},
		{"instanceof", "is_a($o, 'Invoice') instanceof Invoice", `($o instanceof \Invoice) instanceof Invoice`},
		{"cast comparison", "(int) (get_class($o) === 'Invoice')", `(int) ($o instanceof \Invoice)`},
		{"elvis subject", "is_a($p ?: $o, 'Invoice')", `($p ?: $o) instanceof \Invoice`},
		{"not subject", "is_a(!$b, 'Invoice')", `(!$b) instanceof \Invoice`},
		{"cast subject", "is_a((object) $b, 'Invoice')", `(object) $b instanceof \Invoice`},
		{"instanceof subject", "is_a($o instanceof Invoice, 'Invoice')", `($o instanceof Invoice) instanceof \Invoice`},
		{"both", "(bool) is_a($p ?? $o, 'Invoice')", `(bool) (($p ?? $o) instanceof \Invoice)`},
		{"negated loose subject", "get_class($p ?? $o) != 'Invoice'", `!($p ?? $o) instanceof \Invoice`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(head + tc.expr + "; }")
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != 1 || len(got[0].Fixes) != 1 {
				t.Fatalf("findings: got %v, want 1 with a fix", got)
			}
			// The lazy edits may be computed more than once (dry run, then apply).
			if a, b := got[0].Fixes[0].Edits(), got[0].Fixes[0].Edits(); a[0].NewText != b[0].NewText {
				t.Fatalf("edits differ between calls: %q then %q", a[0].NewText, b[0].NewText)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if want := head + tc.fixed + "; }"; string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
