package pregmatchalllayoutmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestAuditRegressions exercises independent valid programs and comment-preservation boundaries.
func TestAuditRegressions(t *testing.T) {
	for _, tc := range []struct {
		src         string
		want, fixes int
	}{
		{"preg_match_all(\"/a.+|b/\",\"aXbbb\",$m,PREG_PATTERN_ORDER,2);echo $m[0][2];", 0, -1},
		{"preg_match_all(\"/(a)/\",\"a\",$m);$m[10][10]=\"new\";", 0, -1},
		// Brace forms whose quantifier meaning depends on the PCRE2 release.
		{"preg_match_all(\"/a{,2}b/\",\"aab ab b\",$m);echo $m[0][2];", 0, -1},
		{"preg_match_all(\"/a{ 2 }b/\",\"aab ab b\",$m);echo $m[0][1];", 0, -1},
		{"preg_match_all(\"/a{1, 2}b/\",\"aab ab b\",$m);echo $m[0][2];", 0, -1},
		{"preg_match_all(\"/a{,}b/\",\"aab ab b\",$m);echo $m[0][2];", 0, -1},
		// Portable quantifiers and plain literal braces keep the proof.
		{"preg_match_all(\"/a{1,2}b/\",\"aab ab b\",$m);echo $m[0][2];", 1, 0},
		{"preg_match_all(\"/a{2}b/\",\"aab ab b\",$m);echo $m[0][1];", 1, 0},
		{"preg_match_all(\"/a{1,}b/\",\"aab ab b\",$m);echo $m[0][2];", 1, 0},
		{"preg_match_all(\"/a{x}b/\",\"a{x}b\",$m);echo $m[0][1];", 1, 0},
		{"preg_match_all(\"/a\\\\{,2}b/\",\"a{,2}b\",$m);echo $m[0][1];", 1, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad regression source: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
			if tc.fixes >= 0 {
				count := 0
				for _, d := range got {
					count += len(d.Fixes)
				}
				if count != tc.fixes {
					t.Fatalf("got %d fixes want %d", count, tc.fixes)
				}
			}
		})
	}
}
