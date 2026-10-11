package propernullcoalescingoperatorusage

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
		{"unset receiver", "<?php function f() { return $u?->m() ?? null; }", 0, ""},
		{"conditional receiver", "<?php function f($c) { if ($c) { $u = g(); } return $u?->m() ?? null; }", 0, ""},
		{"top-level receiver", "<?php $r = $u?->m() ?? null;", 0, ""},
		{"closure import receiver", "<?php $f = function () use ($u, $w) { return $w?->m() ?? null; };", 1, "<?php $f = function () use ($u, $w) { return $w?->m(); };"},
		{"closure local receiver", "<?php $f = function () use ($u) { return $w?->m() ?? null; };", 0, ""},
		{"arrow receiver", "<?php $f = fn () => $u?->m() ?? null;", 0, ""},
		{"parameter receiver", "<?php function f($u) { return $u?->m() ?? null; }", 1, "<?php function f($u) { return $u?->m(); }"},
		{"assigned receiver", "<?php function f() { $u = g(); return $u?->m() ?? null; }", 1, "<?php function f() { $u = g(); return $u?->m(); }"},
		{"property receiver", "<?php function f($o) { return $o->p?->m() ?? null; }", 1, "<?php function f($o) { return $o->p?->m(); }"},
		{"plain call on unset", "<?php function f() { return $u->m() ?? null; }", 1, "<?php function f() { return $u->m(); }"},
		{"callable string", "<?php function f(?callable $c) { return $c ?? 'trim'; }", 0, ""},
		{"callable array", "<?php class A { function f(?callable $c) { return $c ?? [$this, 'f']; } }", 0, ""},
		{"string then callable", "<?php function f(?string $s, callable $c) { return $s ?? $c; }", 0, ""},
		{"closure string", "<?php function f(?Closure $c) { return $c ?? 'trim'; }", 1, ""},
		{"callable int", "<?php function f(?callable $c) { return $c ?? 1; }", 1, ""},
		{"countable array", "<?php function f(?Countable $c) { return $c ?? []; }", 0, ""},
		{"array access array", "<?php function f(?ArrayAccess $c) { return $c ?? []; }", 0, ""},
		{"stringable string", "<?php function f(?Stringable $s) { return $s ?? ''; }", 0, ""},
		{"stringable class string", "<?php class T { function __toString(): string { return ''; } } function f(?T $s) { return $s ?? ''; }", 1, ""},
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
