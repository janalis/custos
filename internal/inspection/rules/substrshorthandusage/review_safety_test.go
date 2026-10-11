package substrshorthandusage

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
		fixed        string
	}{
		{"increment start", "<?php substr($s, $i++, strlen($s) - $i++);", ""},
		{"impure builtin", "<?php substr(array_shift($q), 0, strlen(array_shift($q)) - 1);", ""},
		{"user function", "<?php substr(next_line(), 3, strlen(next_line()) - 1);", ""},
		{"method call", "<?php substr($o->read(), 0, strlen($o->read()) - 2);", ""},
		{"static call", "<?php substr(R::read(), 0, strlen(R::read()) - 2);", ""},
		{"assignment", "<?php substr($s = f, 0, strlen($s = f) - 2);", ""},
		{"new", "<?php substr(new S, 0, strlen(new S) - 2);", ""},
		{"pipe", "<?php substr($s |> read(...), 0, strlen($s |> read(...)) - 2);", ""},
		{"backtick", "<?php substr(`ls`, 0, strlen(`ls`) - 2);", ""},
		{"impure inside pure", "<?php substr(trim($i++), 0, strlen(trim($i++)) - 2);", ""},
		{"write before a later sibling", "<?php substr($a[$i++][$j + 1], 1, strlen($a[$i++][$j + 1]) - 1);", ""},
		{"impure call before a pure one", "<?php substr($a[nx()][strlen('a')], 2, strlen($a[nx()][strlen('a')]) - 2);", ""},
		{"assignment before a later sibling", "<?php substr($a[$i = 1]['0' . ''], 2, strlen($a[$i = 1]['0' . '']) - 2);", ""},
		{"plain binary", "<?php substr($a . $b, 0, strlen($a . $b) - 2);", "<?php substr($a . $b, 0, -2);"},
		{"interpolation", "<?php substr(\"x$a\", 1, strlen(\"x$a\") - 1);", "<?php substr(\"x$a\", 1);"},
		{"pure builtin", "<?php substr(TRIM($s), 0, strlen(trim($s)) - 2);", "<?php substr(TRIM($s), 0, -2);"},
		{"pure start", "<?php mb_substr($s, strlen($h), mb_strlen($s) - strlen($h), 'UTF-8');", "<?php mb_substr($s, strlen($h), null, 'UTF-8');"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != 1 {
				t.Fatalf("findings: got %v, want 1", got)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
		})
	}
}

// A one-argument mb_strlen counts in the internal encoding: with an explicit
// other encoding the computed length is not "up to the end".
func TestReviewEncoding(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for src, want := range map[string]int{
		"<?php substr($s, 1.5, strlen($s) - 1.5);":                   0,
		"<?php mb_substr($s, 0, mb_strlen($s) - 1, '8bit');":         0,
		"<?php mb_substr($s, 1, mb_strlen($s) - 1, 'UTF-16LE');":     0,
		"<?php mb_substr($s, 1, mb_strlen($s) - 1, \"ISO-8859-1\");": 0,
		"<?php mb_substr($s, 1, mb_strlen($s) - 1, $encoding);":      1,
		"<?php mb_substr($s, 1, mb_strlen($s) - 1, \"utf-8\");":      1,
		"<?php mb_substr($s, 0, mb_strlen($s) - 1, 'utf8');":         1,
	} {
		f := syntax.ParseBest("review.php", []byte(src), syntax.Options{Version: phpversion.PHP85})
		if len(f.Errors) != 0 {
			t.Fatalf("invalid probe %s: %v", src, f.Errors)
		}
		if got := e.Analyze(f); len(got) != want {
			t.Errorf("%s: got %v, want %d finding(s)", src, got, want)
		}
	}
}
