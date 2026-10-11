package senselessternaryoperator

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
		{"keyword glue", "<?php function g($a) { return($a===null)?$a:null; }", 1, "<?php function g($a) { return null; }"},
		{"qualified glue", `<?php function g($a) { return($a===\A\B)?$a:\A\B; }`, 1, `<?php function g($a) { return \A\B; }`},
		{"echo glue", "<?php echo($a!==LIMIT)?LIMIT:$a;", 1, "<?php echo LIMIT;"},
		{"variable needs no space", "<?php function g($a) { return($a!==null)?$a:null; }", 1, "<?php function g($a) { return$a; }"},
		{"trailing glue", "<?php $r = LIMIT !== 'a' ? LIMIT : 'a'and $b;", 1, "<?php $r = LIMIT and $b;"},
		{"at end of input", "<?php return $a === 1 ? 1 : $a", 1, "<?php return $a"},
		{"repeated call", "<?php return next($a) === false ? false : next($a);", 0, ""},
		{"user function", "<?php return load($a) === $b ? load($a) : $b;", 0, ""},
		{"namespaced builtin name", "<?php namespace N; function count($a) {} return count($a) === $b ? count($a) : $b;", 0, ""},
		{"pure builtin", "<?php return count($a) === $b ? count($a) : $b;", 1, "<?php return $b;"},
		{"increment", "<?php return $i++ === $b ? $i++ : $b;", 0, ""},
		{"yield", "<?php function g($b) { return (yield 1) === $b ? (yield 1) : $b; }", 0, ""},
		{"negative zero guard", "<?php return $x === 0.0 ? 0.0 : $x;", 0, ""},
		{"signed zero literal", "<?php return $x !== -0.0 ? $x : -0.0;", 0, ""},
		{"integer zero", "<?php return $x === 0 ? 0 : $x;", 1, "<?php return $x;"},
		{"non-zero float", "<?php return $x === 0.5 ? 0.5 : $x;", 1, "<?php return $x;"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
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
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) > len(f.Errors) {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
