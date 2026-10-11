package duplicatematchcondition

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
	const head = "<?php function f($x, $n) { "
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"post increment", head + "$a = 1; $a++; return match($x) { 1 => 'one', $a => 'counter', default => 'other' }; }", 0, ""},
		{"pre decrement", head + "$a = 1; --$a; return match($x) { 1 => 'one', $a => 'counter', default => 'other' }; }", 0, ""},
		{"increment in earlier arm", head + "$a = 1; return match($x) { 1 => 'one', ($a++) + 10 => 'eleven', $a => 'counter', default => 'other' }; }", 0, ""},
		{"by-reference capture", head + "$a = 1; $g = function () use (&$a) { $a = 2; }; $g(); return match($x) { 1 => 'one', $a => 'bumped', default => 'other' }; }", 0, ""},
		{"extract", head + "$a = 1; EXTRACT(['a' => 2]); return match($x) { 1 => 'one', $a => 'extracted', default => 'other' }; }", 0, ""},
		{"catch binding", head + "$a = 1; try { throw $x; } catch (Exception $a) {} return match($x) { 1 => 'one', $a => 'caught', default => 'other' }; }", 0, ""},
		{"variable variable", head + "$a = 1; $$n = 2; return match($x) { 1 => 'one', $a => 'named', default => 'other' }; }", 0, ""},
		{"dynamic condition", head + "return match($x) { 1 => 'one', $$n => 'named', default => 'other' }; }", 0, ""},
		{"eval", head + "$a = 1; eval('$a = 2;'); return match($x) { 1 => 'one', $a => 'eval', default => 'other' }; }", 0, ""},
		{"include", head + "$a = 1; include 'r.php'; return match($x) { 1 => 'one', $a => 'inc', default => 'other' }; }", 0, ""},
		{"reference assignment", head + "$b = 2; $a = &$b; return match($x) { 2 => 'two', $a => 'alias', default => 'other' }; }", 0, ""},
		{"destructuring", head + "[$a] = [1]; return match($x) { 1 => 'one', $a => 'item', default => 'other' }; }", 0, ""},
		{"two assignments", head + "$a = 1; $a = 2; return match($x) { 2 => 'two', $a => 'again', default => 'other' }; }", 0, ""},
		{"copied from a changed variable", head + "$b = 1; $b++; $a = $b; return match($x) { 1 => 'one', $a => 'copy', default => 'other' }; }", 0, ""},
		{"copied from a captured variable", head + "$b = 1; $g = function () use (&$b) { $b = 2; }; $g(); $a = ($b); return match($x) { 1 => 'one', $a => 'copy', default => 'other' }; }", 0, ""},
		{"top level", "<?php $a = 1; echo match($x) { 1 => 'one', $a => 'global', default => 'other' };", 0, ""},
		{"arrow function", "<?php $f = fn($x) => match($x) { 1 => 'one', $x => 'self', default => 'other' };", 0, ""},
		{"by-value capture", head + "$a = 1; $g = function () use ($a) { return $a; }; return match($x) { 1 => 'one', $a => 'same', default => $g() }; }", 1, head + "$a = 1; $g = function () use ($a) { return $a; }; return match($x) { 1 => 'one',  default => $g() }; }"},
		{"single assignment", head + "$a = 1; return match($x) { 1 => 'one', ($a) => 'same', default => 'other' }; }", 1, head + "$a = 1; return match($x) { 1 => 'one',  default => 'other' }; }"},
		{"whole arm repeated", head + "return match($x) { 1, 2 => 'low', 2, 1 => 'never', default => 'other' }; }", 2, head + "return match($x) { 1, 2 => 'low',  default => 'other' }; }"},
		{"whole last arm repeated", head + "return match($x) { 1, 2 => 'low', 2, 1 => 'never' }; }", 2, head + "return match($x) { 1, 2 => 'low' }; }"},
		{"repeat inside one arm", head + "return match($x) { 1 => 'a', 2, 2 => 'b', default => 'other' }; }", 1, ""},
		{"earlier and same-arm repeat", head + "return match($x) { 1 => 'a', 1, 2, 2 => 'b', default => 'other' }; }", 2, ""},
		{"unknown condition keeps arm", head + "return match($x) { 1 => 'a', 1, f() => 'b', default => 'other' }; }", 1, ""},
		{"comment keeps arm", head + "return match($x) { 1, 2 => 'low', 2, /* why */ 1 => 'never', default => 'other' }; }", 2, ""},
		{"nowdoc", head + "return match($x) { <<<'A'\nq\nA => 1, <<<'A'\nq\nA => 2, default => 3 }; }", 0, ""},
		{"interpolated string", head + "return match($x) { \"a$n\" => 1, \"a$n\" => 2, default => 3 }; }", 0, ""},
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
