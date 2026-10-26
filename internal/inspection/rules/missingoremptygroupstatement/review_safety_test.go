package missingoremptygroupstatement

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestReviewSafety pins the exact text of the fix: fixtures compare it with
// whitespace collapsed.
func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, fixed string
	}{
		{"one-line island", `<b <?php if ($a) print "x"?>>`, `<b <?php if ($a) { print "x"; } ?>>`},
		{"one-line island with semicolon", `<b <?php if ($a) echo "x"; ?>>`, `<b <?php if ($a) { echo "x"; } ?>>`},
		{"header on its own line", "<?php\nif ($a)\n    echo 'x' ?>\n", "<?php\nif ($a) {\n    echo 'x';\n} ?>\n"},
		{"plain statement", "<?php\nif ($a) echo 'x';\n", "<?php\nif ($a) {\n    echo 'x';\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			if f := syntax.ParseBest("review.php", src, opt); len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if string(res.Source) != tc.fixed {
				t.Fatalf("fixed source: got %q, want %q", res.Source, tc.fixed)
			}
		})
	}
}
