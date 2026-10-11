package bytelengthusedascharactercount

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
		name, source string
		findings     int
	}{
		{"byte offset", "<?php substr($p, strlen('€'));", 0},
		{"frame length", "<?php pack('N', strlen('héllo'));", 0},
		{"negative offset", "<?php substr($p, -strlen('é'));", 0},
		{"arithmetic", "<?php fseek($h, (strlen('é') + 4) * 2);", 0},
		{"named argument", "<?php substr(string: $p, offset: strlen('é'));", 0},
		{"plain count", "<?php $n = strlen('é');", 1},
		{"statement", "<?php strlen('é');", 1},
		{"arithmetic count", "<?php $n = strlen('é') + 1;", 1},
		{"character consumer", "<?php mb_substr($p, strlen('é'));", 1},
		{"concatenation", "<?php fwrite($h, 'n=' . strlen('é'));", 1},
		{"comparison", "<?php substr($p, strlen('é') > 1);", 1},
		{"logical not", "<?php substr($p, !strlen('é'));", 1},
		{"method", "<?php $o->substr($p, strlen('é'));", 1},
		{"user function", "<?php namespace A; function substr($a, $b) {} substr('x', strlen('é'));", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			for _, d := range got {
				if d.Message != message {
					t.Fatalf("unexpected finding: %+v", d)
				}
			}
		})
	}
}
