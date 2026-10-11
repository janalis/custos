package unnecessaryissetarguments

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
		{"increment index", "<?php isset($a[$i++], $a[$i++]['x']);", 0, ""},
		{"call index", "<?php isset($a[f()], $a[f()]['x']);", 0, ""},
		{"call index reversed", "<?php isset($a[f()]['x'], $a[f()]);", 0, ""},
		{"call container", "<?php isset($o->m()->p, $o->m()->p[0]);", 0, ""},
		{"new container", "<?php isset((new C)->p, (new C)->p[0]);", 0, ""},
		{"assignment index", "<?php isset($a[$i = 1], $a[$i = 1]['x']);", 0, ""},
		// isset() short-circuits: dropping an earlier argument would let the
		// later effect run even when that argument is unset.
		{"pure outer base before the effect", "<?php isset($a, $a[f()]['x']);", 0, ""},
		{"effect in deepest index after the base", "<?php isset($a['k'], $a['k'][$i++]);", 0, ""},
		{"effect between base and access", "<?php isset($a, $b[f()], $a['x']);", 0, ""},
		{"pure outer base after the effect", "<?php isset($a[f()]['x'], $a);", 1, "<?php isset($a[f()]['x']);"},
		{"pure base under effectful one", "<?php isset($a[f()], $a[f()]['x'], $a);", 1, "<?php isset($a[f()], $a[f()]['x']);"},
		{"effect in deepest index before the base", "<?php isset($a['k'][$i++], $a['k']);", 1, "<?php isset($a['k'][$i++]);"},
		{"effect after the access", "<?php isset($a, $a['x'], $b[f()]);", 1, "<?php isset($a['x'], $b[f()]);"},
		{"second access covers safely", "<?php isset($a['x'], $a, $a[f()]);", 1, "<?php isset($a['x'], $a[f()]);"},
		{"variable index", "<?php isset($a[$k], $a[$k]['x']);", 1, "<?php isset($a[$k]['x']);"},
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
