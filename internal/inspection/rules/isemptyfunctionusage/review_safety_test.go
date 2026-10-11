package isemptyfunctionusage

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		style        analysis.ComparisonStyle
		findings     int
		fixed        string
	}{
		{"keyword before call", "<?php function f(): ?DateTime {} function t() { return!empty(f()); }", analysis.StyleRegular, 1, "<?php function f(): ?DateTime {} function t() { return f() !== null; }"},
		{"keyword before qualified call", "<?php function f(): ?DateTime {} function t() { return!empty(\\f()); }", analysis.StyleRegular, 1, "<?php function f(): ?DateTime {} function t() { return \\f() !== null; }"},
		{"keyword before variable", "<?php function t(?DateTime $x) { return!empty($x); }", analysis.StyleRegular, 1, "<?php function t(?DateTime $x) { return$x !== null; }"},
		{"keyword before yoda null", "<?php function t(?DateTime $x) { return!empty($x); }", analysis.StyleYoda, 1, "<?php function t(?DateTime $x) { return null !== $x; }"},
		{"keyword after null", "<?php function t(?DateTime $x, $y) { return [empty($x)and $y, !empty($x)xor$y, empty($x)?1:2]; }", analysis.StyleRegular, 3, "<?php function t(?DateTime $x, $y) { return [$x === null and $y, $x !== null xor$y, $x === null?1:2]; }"},
		{"keyword after yoda variable", "<?php function t(?DateTime $x, $y) { return empty($x)or $y; }", analysis.StyleYoda, 1, "<?php function t(?DateTime $x, $y) { return null === $x or $y; }"},
		{"file start", "<?php !empty(new DateTime);", analysis.StyleRegular, 1, "<?php new DateTime !== null;"},
		{"alias write", "<?php function t(?DateTime $x) { $r = &$x; $r = ''; return empty($x); }", analysis.StyleRegular, 0, ""},
		{"alias target", "<?php function t(?DateTime $x, $o) { $x = &$o; return empty($x); }", analysis.StyleRegular, 0, ""},
		{"alias in array", "<?php function t(?DateTime $x) { $l = [&$x]; return empty($x); }", analysis.StyleRegular, 0, ""},
		{"alias at file scope", "<?php $x = new DateTime; $r = &$x; $r = 0; echo empty($x);", analysis.StyleRegular, 0, ""},
		{"alias of another variable", "<?php function t(?DateTime $x, $o) { $r = &$o; return empty($x); }", analysis.StyleRegular, 1, "<?php function t(?DateTime $x, $o) { $r = &$o; return $x === null; }"},
		{"bcmath number", "<?php function t(?\\BcMath\\Number $n) { return empty($n); }", analysis.StyleRegular, 0, ""},
		{"interface of a falsy class", "<?php function t(?\\Stringable $s, ?\\Countable $c, \\Traversable $t) { return [empty($s), empty($c), !empty($t)]; }", analysis.StyleRegular, 0, ""},
		{"unrelated interface", "<?php function t(?\\JsonSerializable $j) { return empty($j); }", analysis.StyleRegular, 1, "<?php function t(?\\JsonSerializable $j) { return $j === null; }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true, ComparisonStyle: tc.style})
			if err != nil {
				t.Fatal(err)
			}
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
