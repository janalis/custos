package unsafeissetoverarray

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{
		PHP:       phpversion.PHP85,
		EnableAll: true,
		Rules:     map[string]analysis.RuleConfig{"UnSafeIsSetOverArray": {Options: map[string]any{"SUGGEST_TO_USE_NULL_COMPARISON": true}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"comparison operand", "<?php function f($a, $b) { return isset($a) === isset($b); }", 2, "<?php function f($a, $b) { return ($a !== null) === ($b !== null); }"},
		{"cast operand", "<?php function f($a) { return (int) isset($a); }", 1, "<?php function f($a) { return (int) ($a !== null); }"},
		{"inverted arithmetic", "<?php function f($a) { return !isset($a) + 1; }", 1, "<?php function f($a) { return ($a === null) + 1; }"},
		{"logical operand", "<?php function f($a, $b) { return !isset($a) || $b; }", 1, "<?php function f($a, $b) { return $a === null || $b; }"},
		{"maybe undefined", "<?php function f($c) { if ($c) { $v = 1; } return isset($v); }", 1, ""},
		{"variable variable", "<?php function f($c) { return isset($$c); }", 1, ""},
		{"unset parameter", "<?php function f($c) { unset($c); return isset($c); }", 1, ""},
		{"dominating assignment", "<?php function f() { $v = g(); return isset($v); }", 1, "<?php function f() { $v = g(); return $v !== null; }"},
		{"write skipped by goto", "<?php function f($c) { if ($c) { goto done; } $v = 1; done: return isset($v); }", 1, ""},
		{"parameter with goto", "<?php function f($c) { if ($c) { goto done; } $v = 1; done: return isset($c); }", 1, "<?php function f($c) { if ($c) { goto done; } $v = 1; done: return $c !== null; }"},
		{"uninitialized typed", "<?php class C { public int $n; function f() { return isset($this->n); } }", 0, ""},
		{"uninitialized static", "<?php class C { public static int $n; function f() { return isset(self::$n); } }", 0, ""},
		{"private from outside", "<?php class C { private $p; } function f(C $c) { return isset($c->p); }", 0, ""},
		{"magic reader", "<?php class C { public $p; function __isset($n) { return true; } function f() { return isset($this->p); } }", 1, ""},
		{"dynamic unset", "<?php class C { public $p; function f($n) { unset($this->$n); return isset($this->p); } }", 1, ""},
		{"nullable receiver", "<?php class C { public $p; } function f(?C $c) { return isset($c->p); }", 1, ""},
		{"typed with default", "<?php class C { public ?int $p = null; function f() { return isset($this->p); } }", 1, "<?php class C { public ?int $p = null; function f() { return $this->p !== null; } }"},
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
