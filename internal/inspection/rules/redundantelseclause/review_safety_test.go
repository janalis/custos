package redundantelseclause

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
		name, source, fixed string
	}{
		{
			"line comment then code",
			"<?php\nfunction f($a) { if ($a) { return 1; } else {\n    g(); // done\n} h(); return 2; }",
			"<?php\nfunction f($a) { if ($a) { return 1; }\ng(); // done\n h(); return 2; }",
		},
		{
			"hash comment then brace",
			"<?php\nfunction f($a) { if ($a) { return 1; } else {\n    g(); # done\n}}",
			"<?php\nfunction f($a) { if ($a) { return 1; }\ng(); # done\n}",
		},
		{
			"line comment at line end",
			"<?php\nif ($a) { return 1; } else {\n    g(); // done\n}\nh();",
			"<?php\nif ($a) { return 1; }\ng(); // done\nh();",
		},
		{
			"line comment at end of file",
			"<?php\nif ($a) { return 1; } else {\n    g(); // done\n}",
			"<?php\nif ($a) { return 1; }\ng(); // done",
		},
		{
			"block comment then code",
			"<?php\nif ($a) { return 1; } else { g(); /* done */ } h();",
			"<?php\nif ($a) { return 1; }\ng(); /* done */ h();",
		},
		{
			"kept gap comment and empty block",
			"<?php\nif ($a) { return 1; } // why\nelse {} h();",
			"<?php\nif ($a) { return 1; }\n// why\n h();",
		},
		{
			"empty statement and comment",
			"<?php\nif ($a) { return 1; } else { ; // x\n} h();",
			"<?php\nif ($a) { return 1; }; // x\n h();",
		},
		{
			"open tag before brace",
			"<?php\nfunction f($a) { if ($a) { return 1; } else { ?>x<?php }}",
			"<?php\nfunction f($a) { if ($a) { return 1; }\n?>x<?php\n}",
		},
		{
			"else if then code",
			"<?php\nif ($a) { return 1; } else if ($b) { g(); } h();",
			"<?php\nif ($a) { return 1; }\nif ($b) { g(); } h();",
		},
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
			if string(res.Source) != tc.fixed {
				t.Fatalf("fixed source: got %q, want %q", res.Source, tc.fixed)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
