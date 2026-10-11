package passwordcomparedwithfreshhash

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
		{"false", "<?php password_hash($p, PASSWORD_BCRYPT) === false;", 0, ""},
		{"qualified false", `<?php password_hash($p, PASSWORD_BCRYPT) === \FALSE;`, 0, ""},
		{"null left", "<?php null === password_hash($p, PASSWORD_DEFAULT);", 0, ""},
		{"true", "<?php password_hash($p, PASSWORD_DEFAULT) != true;", 0, ""},
		{"integer", "<?php password_hash($p, PASSWORD_DEFAULT) === 12;", 0, ""},
		{"float", "<?php password_hash($p, PASSWORD_DEFAULT) == 1.5;", 0, ""},
		{"empty string", "<?php password_hash($p, PASSWORD_DEFAULT) !== '';", 0, ""},
		{"array", "<?php password_hash($p, PASSWORD_DEFAULT) === [];", 0, ""},
		{"assigned hash", "<?php $h = password_hash($p, PASSWORD_DEFAULT); $h === false;", 0, ""},
		{"assigned sentinel", "<?php $f = false; password_hash($p, PASSWORD_DEFAULT) === $f;", 0, ""},
		{"other constant", "<?php password_hash($p, PASSWORD_DEFAULT) === STORED;", 1, "<?php password_verify($p, STORED);"},
		{"string literal", "<?php password_hash($p, PASSWORD_DEFAULT) === 'x';", 1, "<?php password_verify($p, 'x');"},
		{"stored variable", "<?php password_hash($p, PASSWORD_DEFAULT) === $s;", 1, "<?php password_verify($p, $s);"},
		{"two hashes", "<?php password_hash($p, PASSWORD_DEFAULT) == password_hash($s, PASSWORD_DEFAULT);", 1, ""},
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
