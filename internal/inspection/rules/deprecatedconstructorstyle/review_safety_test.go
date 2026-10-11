package deprecatedconstructorstyle

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
		findings     int
		fixed        string
	}{
		{"plain method on 8.0", "<?php class Filter { function filter($v) { return $v; } }", phpversion.PHP80, 0, ""},
		{"typed method on 8.4", "<?php class Query { function query(): string { return ''; } }", phpversion.PHP84, 0, ""},
		{"legacy on 7.4", "<?php class Filter { function filter($v) {} }", phpversion.PHP74, 1, "<?php class Filter { function __construct($v) {} }"},
		{"legacy on 5.6", "<?php class Ledger { function Ledger() {} }", phpversion.PHP56, 1, "<?php class Ledger { function __construct() {} }"},
		{"return type keeps name", "<?php class Query { function query(): string { return ''; } }", phpversion.PHP74, 1, ""},
		{"abstract return type keeps name", "<?php abstract class Shape { abstract function shape(): void; }", phpversion.PHP71, 1, ""},
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
