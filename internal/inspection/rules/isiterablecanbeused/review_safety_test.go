package isiterablecanbeused

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
		{"increment", "<?php is_array($x[$i++]) || $x[$i++] instanceof Traversable;", 0, ""},
		{"pointer move", "<?php is_array(next($x)) || next($x) instanceof Traversable;", 0, ""},
		{"random", "<?php is_array(rand()) || rand() instanceof Traversable;", 0, ""},
		{"compound assignment", "<?php is_array($a .= 'x') || ($a .= 'x') instanceof Traversable;", 0, ""},
		{"ordinary call", "<?php is_array(nextItem()) || nextItem() instanceof Traversable;", 1, ""},
		{"method call", "<?php is_array($o->get()) || $o->get() instanceof Traversable;", 1, ""},
		{"dynamic call", "<?php is_array($f()) || $f() instanceof Traversable;", 1, ""},
		{"property", "<?php is_array($o->p) || $o->p instanceof Traversable;", 1, ""},
		{"plain variable", "<?php $r = is_array($v) || $v instanceof Traversable;", 1, "<?php $r = is_iterable($v);"},
		{"written qualified", "<?php $r = \\is_array($v) || $v instanceof \\Traversable;", 1, "<?php $r = \\is_iterable($v);"},
		{"namespace", "<?php namespace App; $r = is_array($v) || $v instanceof \\Traversable;", 1, "<?php namespace App; $r = \\is_iterable($v);"},
		{"imported shadow", "<?php use function Foo\\is_iterable; $r = is_array($v) || $v instanceof Traversable;", 1, "<?php use function Foo\\is_iterable; $r = \\is_iterable($v);"},
		{"reversed", "<?php $r = ($v instanceof Traversable) || is_array($v);", 1, "<?php $r = is_iterable($v);"},
		{"chain keeps neighbours", "<?php $r = $a || is_array($v) || $v instanceof Traversable || $b;", 1, "<?php $r = $a || is_iterable($v) || $b;"},
		{"inside and", "<?php $r = $a && (is_array($v) || $v instanceof Traversable);", 1, "<?php $r = $a && (is_iterable($v));"},
		{"split by parentheses", "<?php $r = ($a || is_array($v)) || $v instanceof Traversable;", 1, ""},
		{"comment", "<?php $r = is_array($v) || /* c */ $v instanceof Traversable;", 1, ""},
		{"line comment", "<?php $r = is_array($v) // c\n || $v instanceof Traversable;", 1, ""},
		{"parenthesised argument", "<?php $r = is_array(($v)) || $v instanceof Traversable;", 0, ""},
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
