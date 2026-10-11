package onetimeusevariables

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
		{"by-ref destructuring", "<?php function f() { $p = [1, 2]; [&$a, $b] = $p; $a = 9; return $a + $b; }", 0, ""},
		{"keyed by-ref destructuring", "<?php function f() { $p = g(); ['k' => &$a] = $p; return $a; }", 0, ""},
		{"nested by-ref destructuring", "<?php function f() { $p = g(); [, [$b, &$a]] = $p; return $a; }", 0, ""},
		{"nested by-ref list", "<?php function f() { $p = g(); list($b, list(&$a)) = $p; return $a; }", 0, ""},
		{"nested by-value destructuring", "<?php function f() { $p = g(); [, [$b], list($a)] = $p; return $a; }", 1, "<?php function f() { [, [$b], list($a)] = g(); return $a; }"},
		{"by-ref assignment", "<?php function f($k) { $v = &$this->items[$k]; return $v; }", 0, ""},
		{"by-ref local", "<?php function f() { $v = &$undefined; return $v; }", 0, ""},
		{"match object", "<?php function f($x) { $o = match($x) { default => new B }; return $o->a; }", 1, "<?php function f($x) { return (match($x) { default => new B })->a; }"},
		{"match value", "<?php function f($x) { $o = match($x) { default => 1 }; return $o; }", 1, "<?php function f($x) { return match($x) { default => 1 }; }"},
		{"int object", "<?php function f() { $o = 5; return $o->a; }", 1, "<?php function f() { return (5)->a; }"},
		{"float object", "<?php function f() { $o = 1.5; return $o->a; }", 1, "<?php function f() { return (1.5)->a; }"},
		{"string object", "<?php function f() { $o = 'B'; return $o->a; }", 1, "<?php function f() { return 'B'->a; }"},
		{"interpolated object", "<?php function f($x) { $o = \"b $x\"; return $o->a; }", 1, "<?php function f($x) { return \"b $x\"->a; }"},
		{"nowdoc object", "<?php function f() { $o = <<<'T'\nx\nT; return $o->a; }", 1, "<?php function f() { return (<<<'T'\nx\nT)->a; }"},
		{"heredoc object", "<?php function f($x) { $o = <<<T\n$x\nT; return $o->a; }", 1, "<?php function f($x) { return (<<<T\n$x\nT)->a; }"},
		{"compact names it", "<?php function f() { $p = g(); [$x, $y] = $p; return $x . $y . count(compact('p')); }", 0, ""},
		{"compact names another", "<?php function f() { $p = g(); [$x, $y] = $p; return compact('x', 'y'); }", 1, "<?php function f() { [$x, $y] = g(); return compact('x', 'y'); }"},
		{"compact dynamic name", "<?php function f($n) { $p = g(); [$x, $y] = $p; return compact('x', $n); }", 0, ""},
		{"compact spread", "<?php function f($n) { $p = g(); [$x, $y] = $p; return compact(...$n); }", 0, ""},
		{"compact before", "<?php function f() { $q = compact('p'); $p = g(); [$x, $y] = $p; return $x . $y . count($q); }", 1, "<?php function f() { $q = compact('p'); [$x, $y] = g(); return $x . $y . count($q); }"},
		{"compact before in loop", "<?php function f() { foreach (h() as $i) { $q = compact('p'); $p = g(); [$x, $y] = $p; } return $q; }", 0, ""},
		{"compact in value", "<?php function f() { $p = compact('p'); return $p; }", 1, "<?php function f() { return compact('p'); }"},
		{"compact in value in loop", "<?php function f() { while (h()) { $p = g(compact('p')); [$x] = $p; } return $x; }", 0, ""},
		{"compact in closure", "<?php function f() { $p = g(); [$x] = $p; return function () { return compact('p'); }; }", 1, "<?php function f() { [$x] = g(); return function () { return compact('p'); }; }"},
		{"variable variable", "<?php function f() { $p = g(); [$x, $y] = $p; $n = 'p'; return $x . $y . count($$n); }", 0, ""},
		{"defined vars in finally", "<?php function f() { try { $r = g(); return $r; } finally { log(get_defined_vars()); } }", 0, ""},
		{"include after", "<?php function f() { $p = g(); [$x] = $p; include 'view.php'; }", 0, ""},
		{"eval after", "<?php function f() { $p = g(); [$x] = $p; eval('echo $p[0];'); }", 0, ""},
		{"include as value", "<?php $config = require __DIR__ . '/config.php'; return $config;", 1, "<?php return require __DIR__ . '/config.php';"},
		{"include before at file level", "<?php require 'boot.php'; $app = make(); return $app->run();", 1, "<?php require 'boot.php'; return make()->run();"},
		{"file-level dynamic read", "<?php $p = g(); [$x] = $p; echo $$n;", 0, ""},
		{"other function call", "<?php function f() { $p = g(); [$x] = $p; return h($x); }", 1, "<?php function f() { [$x] = g(); return h($x); }"},
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
