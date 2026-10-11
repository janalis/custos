package arrayrandkeyusedasvalue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	const digit = "<?php function c($x){switch($x){case '0':break;case '1':break;}} "
	for _, tc := range []struct {
		name, source string
		php          phpversion.Version
		findings     int
	}{
		{"keys equal cases", digit + "c(array_rand(['0','1']));", phpversion.PHP85, 0},
		{"swapped keys equal cases", digit + "c(array_rand(['1','0']));", phpversion.PHP85, 0},
		{"keys equal cases before 8.0", digit + "c(array_rand(['0','1']));", phpversion.PHP74, 0},
		{"numeric forms", "<?php function c($x){switch($x){case '0.0':break;case '1e0':break;}} c(array_rand(['0.0','1e0']));", phpversion.PHP85, 0},
		{"key without case", "<?php function c($x){switch($x){case '0':break;case '2':break;}} c(array_rand(['0','2']));", phpversion.PHP85, 1},
		{"text never equals a key", "<?php function c($x){switch($x){case 'red':break;}} c(array_rand(['red']));", phpversion.PHP85, 1},
		{"text equals zero before 8.0", "<?php function c($x){switch($x){case 'red':break;}} c(array_rand(['red']));", phpversion.PHP74, 0},
		{"second key unmatched before 8.0", "<?php function c($x){switch($x){case 'red':break;case 'blue':break;}} c(array_rand(['red','blue']));", phpversion.PHP74, 1},
		{"numeric prefix", "<?php function c($x){switch($x){case '0th':break;case '1st':break;}} c(array_rand(['0th','1st']));", phpversion.PHP85, 1},
		{"numeric prefix before 8.0", "<?php function c($x){switch($x){case '0th':break;case '1st':break;}} c(array_rand(['0th','1st']));", phpversion.PHP74, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			opt := syntax.Options{Version: tc.php}
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
