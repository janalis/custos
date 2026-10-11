package gmpdivisionbyknownzero

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Writes the flow engine does not model must remove the zero proof.
func TestReviewUnmodelledWrites(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$c = 0; $c++; gmp_div_q(1, $c);", 0},
		{"$c = 0; ++$c; gmp_div_q(1, $c);", 0},
		{"$c = 0; $c--; gmp_div_q(1, $c);", 0},
		{"$c = 0; for ($i = 0; $i < 3; $i++, $c++); gmp_div_q(1, $c);", 0},
		{"$c = 0; while ($c++ < 3) {} gmp_div_q(1, $c);", 0},
		{"$c = gmp_init(0); $c++; gmp_div_q(1, $c);", 0},
		{"$a = 0; $a++; $c = $a; gmp_div_q(1, $c);", 0},
		{"$a = 0; $a++; gmp_div_q(1, gmp_init($a));", 0},
		{"$b = 10; $b++; gmp_div_q(1, gmp_init('0', $b));", 0},
		{"$c = 0; try { f(); } catch (E $c) {} gmp_div_q(1, $c);", 0},
		{"$c = 0; $f = function () use (&$c) { $c = 2; }; $f(); gmp_div_q(1, $c);", 0},
		{"$c = 0; extract($row); gmp_div_q(1, $c);", 0},
		{"$c = 0; parse_str($query); gmp_div_q(1, $c);", 0},
		{"function f($row) { $c = 0; extract($row); return gmp_div_q(1, $c); }", 0},
		{"gmp_div_q(1);", 0},
		// Still proven.
		{"$c = 0; gmp_div_q(1, $c);", 1},
		{"$c = 0; $d = $c; $c = $d; gmp_div_q(1, $c);", 1},
		{"$c = 0; $o++; gmp_div_q(1, $c);", 1},
		{"$c = 0; $f = function () use ($c) { extract($c); }; gmp_div_q(1, $c);", 1},
		{"$c = 0; $f = fn () => $c++; gmp_div_q(1, $c);", 1},
		{"function f($row) { $c = 0; g($row); return gmp_div_q(1, $c); }", 1},
		{"extract($row); gmp_div_q(1, 0);", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GmpDivisionByKnownZero"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
