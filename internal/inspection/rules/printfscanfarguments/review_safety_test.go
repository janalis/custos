package printfscanfarguments

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewFormatGrammar(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		php          phpversion.Version
		message      string
	}{
		{"float h", "<?php sprintf('%h', 1.5);", phpversion.PHP85, ""},
		{"float H", "<?php sprintf('%H', 1.5);", phpversion.PHP80, ""},
		{"star width", "<?php sprintf('%*d', 5, 42);", phpversion.PHP85, ""},
		{"star precision", "<?php sprintf('%.*f', 2, 3.14159);", phpversion.PHP85, ""},
		{"star positional", "<?php sprintf('%1$*2$d', 42, 5);", phpversion.PHP85, ""},
		{"star arity", "<?php sprintf('%*d', 42);", phpversion.PHP85, "This call needs 3 argument(s) in total."},
		{"zero plus", "<?php sprintf('%0+5d', 42);", phpversion.PHP85, ""},
		{"minus plus", "<?php sprintf('%-+6d|', 42);", phpversion.PHP56, ""},
		{"h before 8.0", "<?php sprintf('%h', 1.5);", phpversion.PHP74, "Malformed format string."},
		{"star before 8.0", "<?php sprintf('%*d', 5, 42);", phpversion.PHP74, "Malformed format string."},
		{"printf set", "<?php printf('%[a-z]', 'x');", phpversion.PHP85, "Malformed format string."},
		{"fprintf star", "<?php fprintf($fh, '%*d', 5, 42);", phpversion.PHP85, ""},
		{"scanf i n", "<?php sscanf($s, '%i %s%n', $a, $b, $n);", phpversion.PHP85, ""},
		{"scanf h", "<?php sscanf($s, '%hd', $h);", phpversion.PHP56, ""},
		{"scanf D", "<?php sscanf($s, '%D', $d);", phpversion.PHP85, ""},
		{"scanf percent in set", "<?php sscanf($s, '%[^%]%%%s', $x, $y);", phpversion.PHP85, ""},
		{"scanf set arity", "<?php sscanf($s, '%[^%]%%%s', $x);", phpversion.PHP85, "This call needs 4 argument(s) in total."},
		{"scanf suppressed set", "<?php fscanf($fh, '%*[^,],%d', $d);", phpversion.PHP85, ""},
		{"scanf open set", "<?php sscanf($s, '%[a-z', $w);", phpversion.PHP85, "Malformed format string."},
		{"scanf precision", "<?php sscanf($s, '%.2f', $f);", phpversion.PHP85, "Malformed format string."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: tc.php})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if tc.message == "" {
				if len(got) != 0 {
					t.Fatalf("findings: got %v, want none", got)
				}
				return
			}
			if len(got) != 1 || got[0].Message != tc.message {
				t.Fatalf("findings: got %v, want %q", got, tc.message)
			}
		})
	}
}
