package cascadestringreplacement

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
	const (
		first  = "$s = str_replace('a', 'b', $s); "
		second = "return str_replace('c', 'd', $s);"
		merged = "return str_replace(array('a', 'c'), array('b', 'd'), $s);"
	)
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string // "" when the source must be left as it is
	}{
		// a return entry drops the assignment: the variable must be a plain local
		{"local", "<?php function f($s, &$o) { " + first + second + " }", 1, "<?php function f($s, &$o) { " + merged + " }"},
		{"top level", "<?php " + first + second, 0, ""},
		{"by-reference parameter", "<?php function f(&$s) { " + first + second + " }", 0, ""},
		{"global", "<?php function f() { global $s; " + first + second + " }", 0, ""},
		{"static", "<?php function f() { static $s = 'ac'; " + first + second + " }", 0, ""},
		{"closure reference import", "<?php $f = function () use (&$s) { " + first + second + " };", 0, ""},
		{"closure value import", "<?php $f = function () use ($s) { " + first + second + " };", 1, "<?php $f = function () use ($s) { " + merged + " };"},
		{"nested reference import", "<?php function f($s) { $g = function () use (&$s) {}; " + first + second + " }", 0, ""},
		{"reference alias", "<?php function f($s, $o) { $o->p = &$s; " + first + second + " }", 0, ""},
		{"reference element", "<?php function f($s, $o) { $o->p = [&$s]; " + first + second + " }", 0, ""},
		{"reference foreach", "<?php function f($s, $xs) { foreach ($xs as &$s) { " + first + second + " } }", 0, ""},
		{"value foreach", "<?php function f($s, $xs) { foreach ($xs as $s) { " + first + second + " } }", 1, "<?php function f($s, $xs) { foreach ($xs as $s) { " + merged + " } }"},
		{"finally", "<?php function f($s) { try { " + first + second + " } finally { echo $s; } }", 0, ""},
		{"try without finally", "<?php function f($s) { try { " + first + second + " } catch (E $e) {} }", 1, "<?php function f($s) { try { " + merged + " } catch (E $e) {} }"},
		{"assignment entry keeps the write", "<?php function f(&$s) { " + first + "$s = str_replace('c', 'd', $s); }", 1, "<?php function f(&$s) { $s = str_replace(array('a', 'c'), array('b', 'd'), $s); }"},

		// arguments that may be a string or an array are not merged
		{"untyped search", "<?php function f($s, $x, $y) { $s = str_replace($x, $y, $s); $s = str_replace(\"\\n\", ' ', $s); }", 1, ""},
		{"array or string", "<?php function f(string $s, array|string $x) { return str_replace('x', 'y', str_replace($x, '-', $s)); }", 1, ""},
		{"nullable array", "<?php function f(string $s, ?array $x) { $s = str_replace('x', 'y', $s); $s = str_replace($x, [], $s); }", 1, ""},
		{"typed string", "<?php function f($s, string $x, int $y) { $s = str_replace($x, $y, $s); $s = str_replace('c', 'd', $s); }", 1, "<?php function f($s, string $x, int $y) { $s = str_replace(array($x, 'c'), array($y, 'd'), $s); }"},
		{"boxed variable", "<?php function f($s, string $x) { $s = str_replace(['a'], [$x], $s); $s = str_replace('c', '-', $s); }", 1, "<?php function f($s, string $x) { $s = str_replace(array('a', 'c'), array($x, '-'), $s); }"},

		// search and replace lists must stay paired
		{"short replace list", "<?php function f($s) { $s = str_replace(['a', 'b'], ['1'], $s); $s = str_replace('c', '3', $s); }", 1, ""},
		{"long replace list", "<?php function f($s) { return str_replace('c', '3', str_replace(['a'], ['1', '2'], $s)); }", 1, ""},
		{"empty replace list", "<?php function f($s) { $s = str_replace(['a', 'b'], [], $s); $s = str_replace(['c', 'd'], ['3', '4'], $s); }", 1, ""},
		{"empty search list", "<?php function f($s) { $s = str_replace([], 'x', $s); $s = str_replace('c', '3', $s); }", 1, ""},
		{"short list in the latest call", "<?php function f($s) { $s = str_replace('c', '3', $s); $s = str_replace(['a', 'b'], ['1'], $s); }", 1, "<?php function f($s) { $s = str_replace(array('c', 'a', 'b'), array('3', '1'), $s); }"},
		{"boxed replace of a longer search", "<?php function f($s) { $s = str_replace(['a', 'b'], ['-'], $s); $s = str_replace('c', '-', $s); }", 1, ""},
		{"boxed replace in the latest call", "<?php function f($s) { $s = str_replace('c', '-', $s); $s = str_replace(['a', 'b'], ['-'], $s); }", 1, "<?php function f($s) { $s = str_replace(array('c', 'a', 'b'), array('-', '-'), $s); }"},
		{"spread search, string replace", "<?php function f($s, array $m) { $s = str_replace('a', 'b', $s); $s = str_replace([...$m], 'y', $s); }", 1, ""},
		{"spread lists in the latest call", "<?php function f($s, array $m, array $n) { $s = str_replace('a', 'b', $s); $s = str_replace([...$m], [...$n], $s); }", 1, "<?php function f($s, array $m, array $n) { $s = str_replace(array('a', ...$m), array('b', ...$n), $s); }"},
		{"array replace of a list", "<?php function f($s, array $n) { $s = str_replace(['a', 'b'], $n, $s); $s = str_replace('c', 'd', $s); }", 1, ""},
		{"array replace in the latest call", "<?php function f($s, array $n) { $s = str_replace('c', 'd', $s); $s = str_replace(['a', 'b'], $n, $s); }", 1, "<?php function f($s, array $n) { $s = str_replace(array('c', 'a', 'b'), array('d', ...array_values($n)), $s); }"},

		// operands keep their evaluation order
		{"two effects", "<?php function f($s) { $s = str_replace(chr(10), 'b', $s); $s = str_replace(chr(13), 'x', $s); }", 1, ""},
		{"subject writes an argument", "<?php function f($s, string $k) { $s = str_replace('a', 'b', trim($k = 'c')); $s = str_replace($k, 'x', $s); }", 1, ""},
		{"subject writes a replacement", "<?php function f($s, string $k) { $s = str_replace('a', 'b', trim($k = 'c')); $s = str_replace('x', $k, $s); }", 1, ""},
		{"subject apart from arguments", "<?php function f($in, string $k) { $s = str_replace('a', 'b', trim($in)); $s = str_replace($k, 'x', $s); }", 1, "<?php function f($in, string $k) { $s = str_replace(array('a', $k), array('b', 'x'), trim($in)); }"},
		{"effect next to a variable", "<?php function f($in, string $k) { $s = str_replace(strtolower($k), 'b', $in); $s = str_replace($k, 'x', $s); }", 1, ""},
		{"effect after a plain subject", "<?php function f($in) { $s = str_replace('a', 'b', $in); $s = str_replace(chr(10), ' ', $s); }", 1, "<?php function f($in) { $s = str_replace(array('a', chr(10)), array('b', ' '), $in); }"},
		{"effect after a subject with effects", "<?php function f($in) { $s = str_replace('a', 'b', trim($in)); $s = str_replace(chr(10), ' ', $s); }", 1, ""},
		{"effect reading the subject's variable", "<?php function f(string $in) { $s = str_replace('a', 'b', $in); $s = str_replace(strtolower($in = 'q'), ' ', $s); }", 1, ""},
		{"effect after a dynamic subject", "<?php function f($n) { $s = str_replace('a', 'b', $$n); $s = str_replace(chr(10), ' ', $s); }", 1, ""},
		{"effect in the first call", "<?php function f($in) { $s = str_replace(chr(10), ' ', trim($in)); $s = str_replace('a', 'b', $s); }", 1, "<?php function f($in) { $s = str_replace(array(chr(10), 'a'), array(' ', 'b'), trim($in)); }"},
		{"nested effect", "<?php function f($in) { return str_replace(chr(10), ' ', str_replace('a', 'b', $in)); }", 1, "<?php function f($in) { return str_replace(array('a', chr(10)), array('b', ' '), $in); }"},
		{"nested subject with effects", "<?php function f($in, string $k) { return str_replace($k, ' ', str_replace('a', 'b', trim($in))); }", 1, "<?php function f($in, string $k) { return str_replace(array('a', $k), array('b', ' '), trim($in)); }"},
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
				t.Fatalf("fixed source:\n got %s\nwant %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
