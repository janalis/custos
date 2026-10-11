package validatedintegerzerorejected

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A report implies the `=== false` fix, so the rule must stay silent whenever
// a rejected value is not reported as false or the result is not an integer.
func TestReviewFailureValueSafety(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"default zero", "!filter_var($v, FILTER_VALIDATE_INT, ['options' => ['default' => 0]]);", 0},
		{"default null", "!filter_var($v, FILTER_VALIDATE_INT, ['options' => ['default' => null]]);", 0},
		{"default with bound", "!filter_var($v, FILTER_VALIDATE_INT, ['options' => ['min_range' => 0, 'default' => 5]]);", 0},
		{"named options default", "!filter_var($v, options: ['options' => ['default' => 1]], filter: FILTER_VALIDATE_INT);", 0},
		{"require array", "!filter_var($v, FILTER_VALIDATE_INT, FILTER_REQUIRE_ARRAY);", 0},
		{"force array", "!filter_var($v, FILTER_VALIDATE_INT, FILTER_FORCE_ARRAY);", 0},
		{"combined force array", "!filter_var($v, FILTER_VALIDATE_INT, FILTER_FLAG_ALLOW_HEX | FILTER_FORCE_ARRAY);", 0},
		{"flags entry force array", "!filter_var($v, FILTER_VALIDATE_INT, ['flags' => FILTER_FORCE_ARRAY]);", 0},
		{"flags entry combined", "!filter_var($v, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX | FILTER_REQUIRE_ARRAY]);", 0},
		{"literal array flag", "!filter_var($v, FILTER_VALIDATE_INT, 16777216);", 0},
		{"unknown flags entry", "!filter_var($v, FILTER_VALIDATE_INT, ['flags' => $f]);", 0},
		{"unknown inner options", "!filter_var($v, FILTER_VALIDATE_INT, ['options' => $o]);", 0},
		{"scalar flag", "!filter_var($v, FILTER_VALIDATE_INT, FILTER_REQUIRE_SCALAR);", 1},
		{"hex flag", "!filter_var($v, FILTER_VALIDATE_INT, FILTER_FLAG_ALLOW_HEX);", 1},
		{"flags entry hex", "!filter_var($v, FILTER_VALIDATE_INT, ['flags' => FILTER_FLAG_ALLOW_HEX]);", 1},
		{"bounds around zero", "!filter_var($v, FILTER_VALIDATE_INT, ['options' => ['min_range' => -1, 'max_range' => 1]]);", 1},
		{"empty options", "!filter_var($v, FILTER_VALIDATE_INT, []);", 1},
		{"default outside options entry", "!filter_var($v, FILTER_VALIDATE_INT, ['default' => 3]);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ValidatedIntegerZeroRejected"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
