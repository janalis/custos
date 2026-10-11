package scandirfailureiterated

import (
	"testing"

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
		name, body string
		findings   int
	}{
		{"falsy exit", "if (!$fs) { return; } $n = 0; foreach ($fs as $f) {}", 0},
		{"truthy branch", "if ($fs) { foreach ($fs as $f) {} }", 0},
		{"array branch", "if (is_array($fs)) { foreach ($fs as $f) {} }", 0},
		{"not array exit", "if (!is_array($fs)) { return; } foreach ($fs as $f) {}", 0},
		{"loose exit", "if ($fs == false) { return; } foreach ($fs as $f) {}", 0},
		{"empty exit", "if (empty($fs)) { return; } foreach ($fs as $f) {}", 0},
		{"either exit", "if (!$fs || $c) { return; } foreach ($fs as $f) {}", 0},
		{"both branch", "if ($c && $fs) { foreach ($fs as $f) {} }", 0},
		{"conjunction exit", "if (!$fs && $c) { return; } foreach ($fs as $f) {}", 1},
		{"disjunction branch", "if ($fs || $c) { foreach ($fs as $f) {} }", 1},
		{"assertion", "assert($fs !== false); foreach ($fs as $f) {}", 1},
		{"fallthrough guard", "if (!$fs) { echo 0; } foreach ($fs as $f) {}", 1},
		{"rewritten in loop", "if (!$fs) { return; } while ($c) { foreach ($fs as $f) {} $fs = scandir($c); }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php function probe($d, $c) { $fs = scandir($d); "+tc.body+" }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
