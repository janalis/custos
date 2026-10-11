package duplicatearraykeys

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewVersionedEscapes(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		php          phpversion.Version
		findings     int
	}{
		{"codepoint decoded", `<?php $a = array("\u{41}" => 1, 'A' => 2);`, phpversion.PHP70, 1},
		{"codepoint spelled out", `<?php $a = array("\u{41}" => 1, '\u{41}' => 2);`, phpversion.PHP70, 0},
		{"codepoint literal before 7.0", `<?php $a = array("\u{41}" => 1, 'A' => 2);`, phpversion.PHP56, 0},
		{"codepoint text before 7.0", `<?php $a = array("\u{41}" => 1, '\u{41}' => 2);`, phpversion.PHP56, 1},
		{"escaped backslash before 7.0", `<?php $a = array("\\u{41}" => 1, "\u{41}" => 2);`, phpversion.PHP56, 1},
		{"other escapes around it", `<?php $a = array("\x41\u{42}\n" => 1, 'A\u{42}' . "\n" => 2, "A\\u{42}\12" => 3);`, phpversion.PHP56, 1},
		{"binary prefix", `<?php $a = array(b"\u{41}" => 1, 'A' => 2, '\u{41}' => 3);`, phpversion.PHP56, 1},
		{"single quotes holding a double quote", `<?php $a = array('a"\u{41}' => 1, "a\"\u{41}" => 2);`, phpversion.PHP56, 1},
		{"single quotes opening with a double quote", `<?php $a = array('"\u{41}"' => 1, '\u{41}"' => 2, "\"\u{41}\"" => 3);`, phpversion.PHP56, 1},
		{"escape byte from 5.4", `<?php $a = array("\e" => 1, "\x1b" => 2, '\e' => 3);`, phpversion.PHP54, 1},
		{"escape byte unknown in 5.3", `<?php $a = array("\e" => 1, "\x1b" => 2);`, phpversion.PHP53, 0},
		{"escape byte text in 5.3", `<?php $a = array("\e" => 1, '\e' => 2);`, phpversion.PHP53, 1},
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
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
