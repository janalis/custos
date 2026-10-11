package referencingobjects

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
		{"swap", "<?php function f(N &$a, N &$b) { list($a, $b) = array($b, $a); }", 0, ""},
		{"short pattern", "<?php function f(N &$a) { [$a, $x] = [$a->next, 1]; }", 0, ""},
		{"foreach value", "<?php function f(N &$a) { foreach ([$a->next] as $a) {} }", 0, ""},
		{"alias", "<?php function f(N &$a) { $r = &$a; $r = new N; }", 0, ""},
		{"reference item", "<?php function f(N &$a) { $x = [&$a]; $x[0] = new N; }", 0, ""},
		{"catch", "<?php function f(N &$a) { try { g(); } catch (E $a) {} }", 0, ""},
		{"variadic element", "<?php function f(N &...$rest) { $rest[0] = new N; }", 0, ""},
		{"promoted", "<?php final class H { public function __construct(public N &$n) {} }", 0, ""},
		{"reference return", "<?php function &f(N &$a) { return $a; }", 0, ""},
		{"value return", "<?php function f(N &$a) { return $a; }", 1, "<?php function f(N $a) { return $a; }"},
		{"array literal", "<?php function f(N &$a) { return [$a]; }", 1, "<?php function f(N $a) { return [$a]; }"},
		{"unused variadic", "<?php function f(N &...$rest) {}", 1, "<?php function f(N ...$rest) {}"},
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
