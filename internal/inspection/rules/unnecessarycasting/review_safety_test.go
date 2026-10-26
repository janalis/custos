package unnecessarycasting

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
		{"self assignment", "<?php function f(string $k) {\n    $k = (string) $k;\n    return $k;\n}\n", 1, "<?php function f(string $k) {\n    return $k;\n}\n"},
		{"self assignment sharing its line", "<?php function f(string $k) { $k = (string) ($k); return $k; }", 1, "<?php function f(string $k) {  return $k; }"},
		{"self assignment with comment", "<?php function f(string $k) {\n    $k = (string) $k; // keep\n    return $k;\n}\n", 1, ""},
		{"self assignment with inner comment", "<?php function f(string $k) { $k = /* s */ (string) $k; }", 1, ""},
		{"self assignment at end of file", "<?php function f(string $k) { $k = (string) $k; }", 1, "<?php function f(string $k) {  }"},
		{"other target", "<?php function f(string $k) {\n    $j = (string) $k;\n    return $j;\n}\n", 1, "<?php function f(string $k) {\n    $j = $k;\n    return $j;\n}\n"},
		{"nested assignment", "<?php function f(string $k) { return $k = (string) $k; }", 1, ""},
		{"by-reference function", "<?php function fill(&$o) { $o = 'x'; } function f() { $n = 0; fill($n); return (int) $n; }", 0, ""},
		{"unresolved method after cast", "<?php function f($o, int $n) { $c = (int) $n; $o->missing($n); return $c; }", 0, ""},
		{"unresolved function", "<?php function f() { $n = 0; missing($n); return (int) $n; }", 0, ""},
		{"callable variable", "<?php function f($g) { $n = 0; $g($n); return (int) $n; }", 0, ""},
		{"variadic out parameter", "<?php function f() { $n = 0; sscanf('abc', '%s', $n); return (int) $n; }", 0, ""},
		{"settype", "<?php function f() { $n = 0; settype($n, 'string'); return (int) $n; }", 0, ""},
		{"array passed by reference", "<?php function f($in) { $a = [1]; array_unshift($a, $in); return (int) $a[0]; }", 0, ""},
		{"by-value function", "<?php function show($v) {} function f() { $n = 0; show($n); return (int) $n; }", 1, "<?php function show($v) {} function f() { $n = 0; show($n); return $n; }"},
		{"reference alias", "<?php function f(int $n) { $r = &$n; $r = 'x'; return (int) $n; }", 0, ""},
		{"element alias", "<?php function f($in) { $a = ['k' => 1]; $b = &$a['k']; $b = $in; return (int) $a['k']; }", 0, ""},
		{"foreach by reference", "<?php function f($in) { $a = ['k' => 1]; foreach ($a as &$v) { $v = $in; } return (int) $a['k']; }", 0, ""},
		{"alias before a loop", "<?php function f($in) { $n = 0; $r = &$n; $r = $in; foreach ([1] as $x) {} return (int) $n; }", 0, ""},
		{"by-reference call before a loop", "<?php function fill(&$o) { $o = 'x'; } function f() { $n = 0; fill($n); foreach ([1] as $x) {} return (int) $n; }", 0, ""},
		{"by-reference call before a sibling call", "<?php function fill(&$o) { $o = 'x'; } function f() { $n = 0; $z = [fill($n), strlen('x')]; return (int) $n; }", 0, ""},
		{"global statement", "<?php $t = 0; function g() { global $t; $t = 'x'; } g(); return (int) $t;", 0, ""},
		{"globals element", "<?php $t = 0; function g() { $GLOBALS['t'] = 'x'; } g(); return (int) $t;", 0, ""},
		{"plain top level", "<?php $t = 0; return (int) $t;", 1, "<?php $t = 0; return $t;"},
		{"minus glue", "<?php function f(int $b) { return 10 -(int)-$b; }", 1, "<?php function f(int $b) { return 10 - -$b; }"},
		{"plus glue", "<?php function f(int $b) { return 1 +(int)+$b; }", 1, "<?php function f(int $b) { return 1 + +$b; }"},
		{"dot glue", "<?php function f(string $s) { return $s.(string)2; }", 1, "<?php function f(string $s) { return $s. 2; }"},
		{"keyword glue", "<?php function f(array $i) { return(int)count($i); }", 1, "<?php function f(array $i) { return count($i); }"},
		{"no separator needed", "<?php function f(int $b) { return [(int)$b, (int) $b]; }", 2, "<?php function f(int $b) { return [$b, $b]; }"},
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
