package codepointslicesplitsgrapheme

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
		{"mb_substr(\"\\n\\u{0301}x\",1,1,\"UTF-8\");", 0, -1},
		// Format and separator characters that end a cluster like a control.
		{"mb_substr(\"\\u{200B}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{2028}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{2029}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{AD}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{9F}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{FEFF}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{E0001}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		{"mb_substr(\"\\u{E0FFF}\\u{301}x\",0,1,\"UTF-8\");", 0, -1},
		// Format characters that stay inside the cluster keep the report.
		{"mb_substr(\"\\u{200C}\\u{301}x\",0,1,\"UTF-8\");", 1, 0},
		{"mb_substr(\"\\u{200D}\\u{301}x\",0,1,\"UTF-8\");", 1, 0},
		{"mb_substr(\"\\u{600}\\u{301}x\",0,1,\"UTF-8\");", 1, 0},
		{"mb_substr(\"a\\u{E0061}\\u{301}x\",0,2,\"UTF-8\");", 1, 0},
		{"mb_substr(\"\\u{A0}\\u{301}x\",0,1,\"UTF-8\");", 1, 0},
		{"mb_substr(\"\\u{10FFFF}\\u{301}x\",0,1,\"UTF-8\");", 1, 0},
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
