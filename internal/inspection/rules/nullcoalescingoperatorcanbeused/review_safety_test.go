package nullcoalescingoperatorcanbeused

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
		{"reference hook", "<?php class C { public array $a; public mixed $p { &get { if (isset($this->a['k'])) { return $this->a['k']; } return $this->a; } } }", 0, ""},
		{"value hook", "<?php class C { public array $a; public mixed $p { get { if (isset($this->a['k'])) { return $this->a['k']; } return null; } } }", 1, "<?php class C { public array $a; public mixed $p { get { return $this->a['k'] ?? null; } } }"},
		// The probe is evaluated twice by the original, once by `??`: report only.
		{"null call", "<?php $r = nxt() !== null ? nxt() : 'none';", 1, ""},
		{"null method", "<?php $r = null === $q->pop() ? 'none' : $q->pop();", 1, ""},
		{"isset call index", "<?php $r = isset($a[nxt()]) ? $a[nxt()] : 'none';", 1, ""},
		{"isset increment", "<?php $r = isset($a[++$i]) ? $a[++$i] : 'none';", 1, ""},
		{"key exists call", "<?php $r = array_key_exists(nxt(), $a) ? $a[nxt()] : null;", 1, ""},
		{"empty call base", "<?php $r = !empty(box()) ? box()->size : null;", 1, ""},
		{"if return call", "<?php function f() { if (nxt() !== null) { return nxt(); } return 'none'; }", 1, ""},
		{"if else assign call", "<?php if (isset($a[nxt()])) { $v = $a[nxt()]; } else { $v = 1; }", 1, ""},
		{"closure index", "<?php $r = isset($a[$k]) ? $a[$k] : fn() => nxt();", 1, "<?php $r = $a[$k] ?? fn() => nxt();"},
		{"fallback call", "<?php $r = isset($a[$k]) ? $a[$k] : nxt();", 1, "<?php $r = $a[$k] ?? nxt();"},
		// Constant, property and named-argument names are case-sensitive
		// even when spelled like a keyword.
		{"keyword constant", "<?php $r = Foo::ARRAY !== null ? Foo::Array : 1;", 0, ""},
		{"keyword constant reversed", "<?php $r = null === Foo::DEFAULT ? 'x' : Foo::Default;", 0, ""},
		{"keyword property", "<?php $r = isset($o->List) ? $o->list : 1;", 0, ""},
		{"keyword named argument", "<?php $r = f(array: 1) !== null ? f(ARRAY: 1) : 1;", 0, ""},
		{"keyword constant target", "<?php if (isset($a)) { $v[Foo::ARRAY] = $a; } else { $v[Foo::Array] = 1; }", 0, ""},
		{"class keyword", "<?php $r = isset($a[foo::CLASS]) ? $a[Foo::class] : 1;", 1, "<?php $r = $a[Foo::class] ?? 1;"},
		{"class keyword then constant", "<?php $r = isset($a[foo::CLASS][Foo::ARRAY]) ? $a[Foo::class][Foo::Array] : 1;", 0, ""},
		{"same constant", "<?php $r = isset($a[foo::ARRAY]) ? $a[Foo::ARRAY] : 1;", 1, "<?php $r = $a[Foo::ARRAY] ?? 1;"},
		{"dynamic names", "<?php $r = isset($o->{$p}[foo::{$c}]) ? $o->{$p}[Foo::{$c}] : 1;", 1, "<?php $r = $o->{$p}[Foo::{$c}] ?? 1;"},
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
