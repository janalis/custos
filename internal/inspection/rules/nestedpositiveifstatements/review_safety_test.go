package nestedpositiveifstatements

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
		{"negated assignment", "<?php if (!$r = get()) { if ($s) { a(); } }", 1, "<?php if ((!$r = get()) && $s) { a(); }"},
		{"silenced assignment", "<?php if (@$u = $v) { if ($w) { a(); } }", 1, "<?php if ((@$u = $v) && $w) { a(); }"},
		{"cast assignment", "<?php if ((bool) $u = $v) { if ($w) { a(); } }", 1, "<?php if (((bool) $u = $v) && $w) { a(); }"},
		{"plain negation", "<?php if (!$u) { if ($w) { a(); } }", 1, "<?php if (!$u && $w) { a(); }"},
		{"print", "<?php if (print 'p') { if ($w) { a(); } }", 1, "<?php if ((print 'p') && $w) { a(); }"},
		{"include", "<?php if (include 'f.php') { if ($w) { a(); } }", 1, "<?php if ((include 'f.php') && $w) { a(); }"},
		{"throw", "<?php if (throw $e) { if ($w) { a(); } }", 1, "<?php if ((throw $e) && $w) { a(); }"},
		{"yield", "<?php function g($v, $w) { if (yield $v) { if ($w) { a(); } } }", 1, "<?php function g($v, $w) { if ((yield $v) && $w) { a(); } }"},
		{"yield from", "<?php function g($v, $w) { if (yield from $v) { if ($w) { a(); } } }", 1, "<?php function g($v, $w) { if ((yield from $v) && $w) { a(); } }"},
		{"arrow function", "<?php if (fn() => $v) { if ($w) { a(); } }", 1, "<?php if ((fn() => $v) && $w) { a(); }"},
		{"and chain ending in assignment", "<?php if ($a && $r = get()) { if ($s) { a(); } }", 1, "<?php if (($a && $r = get()) && $s) { a(); }"},
		{"and chain", "<?php if ($a && $b) { if ($s) { a(); } }", 1, "<?php if ($a && $b && $s) { a(); }"},
		{"clone of assignment", "<?php if (clone $u = $v) { if ($w) { a(); } }", 1, "<?php if ((clone $u = $v) && $w) { a(); }"},
		{"clone", "<?php if (clone $u) { if ($w) { a(); } }", 1, "<?php if (clone $u && $w) { a(); }"},
		{"clone call form", "<?php if (clone($u)) { if ($w) { a(); } }", 1, "<?php if (clone($u) && $w) { a(); }"},
		{"ternary child", "<?php if ($u) { if ($a ? $b : $c) { a(); } }", 1, "<?php if ($u && ($a ? $b : $c)) { a(); }"},
		{"short ternary child", "<?php if ($u) { if ($a ?: $c) { a(); } }", 1, "<?php if ($u && ($a ?: $c)) { a(); }"},
		{"coalesce child", "<?php if ($u) { if ($a ?? $b) { a(); } }", 1, "<?php if ($u && ($a ?? $b)) { a(); }"},
		{"xor child", "<?php if ($u) { if ($a xor $b) { a(); } }", 1, "<?php if ($u && ($a xor $b)) { a(); }"},
		{"and child", "<?php if ($u) { if ($a and $b) { a(); } }", 1, "<?php if ($u && $a and $b) { a(); }"},
		{"assignment child", "<?php if ($u) { if ($x = $v) { a(); } }", 1, "<?php if ($u && ($x = $v)) { a(); }"},
		{"equality", "<?php if ($a == $b) { if ($w) { a(); } }", 1, "<?php if ($a == $b && $w) { a(); }"},
		{"concatenation", "<?php if ($a . $b) { if ($w) { a(); } }", 1, "<?php if ($a . $b && $w) { a(); }"},
		{"equality ending in assignment", "<?php if ($a == $r = get()) { if ($w) { a(); } }", 1, "<?php if (($a == $r = get()) && $w) { a(); }"},
		{"coalesce parent", "<?php if ($a ?? $b) { if ($w) { a(); } }", 1, "<?php if (($a ?? $b) && $w) { a(); }"},
		{"and keyword parent", "<?php if ($a and $b) { if ($w) { a(); } }", 1, "<?php if (($a and $b) && $w) { a(); }"},
		{"dedented body", "<?php\nif ($u) {\n    // note\n    if ($w) {\n        a();\n        b();\n    }\n}\n", 1, "<?php\nif ($u && $w) {\n    // note\n    a();\n    b();\n}\n"},
		{"dedented else if", "<?php\nif ($u) {\n    a();\n} else {\n    if ($w) {\n        b();\n    }\n}\n", 1, "<?php\nif ($u) {\n    a();\n} else if ($w) {\n    b();\n}\n"},
		{"dedented else if with comment", "<?php\nif ($u) {\n    a();\n} else {\n    // why\n    if ($w\n        || $v) {\n        b();\n    }\n}\n", 1, "<?php\nif ($u) {\n    a();\n} else if ($w\n    || $v) {\n    // why\n    b();\n}\n"},
		{"multi-line string kept", "<?php\nif ($u) {\n    if ($w) {\n        a('x\n        y');\n    }\n}\n", 1, "<?php\nif ($u && $w) {\n        a('x\n        y');\n    }\n"},
		{"colon child in if", "<?php if ($u) { if ($v): a(); endif; }", 0, ""},
		{"colon child with else in if", "<?php if ($u) { if ($v): a(); else: b(); endif; } else { b(); }", 0, ""},
		{"colon child in else", "<?php if ($u) { a(); } else { if ($v): b(); endif; }", 1, "<?php if ($u) { a(); } else if ($v): b(); endif;"},
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
