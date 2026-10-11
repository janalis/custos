package nestedassignmentsusage

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
		fixed        string
	}{
		{"close tag", "<?php $a = $b = 1 ?>\n<p></p>", "<?php $b = 1;\n$a = 1; ?>\n<p></p>"},
		{"close tag braces", "<?php if ($c) $a = $b = 1 ?>\n<p></p>", "<?php if ($c) {\n    $b = 1;\n    $a = 1;\n} ?>\n<p></p>"},
		{"typed property", "<?php $shown = $this->ratio = 0;", ""},
		{"static property", "<?php $n = self::$total = '7';", ""},
		{"string offset", "<?php $t = $s[0] = 'xyz';", ""},
		{"middle property", "<?php $a = $this->ratio = $b = f();", ""},
		{"null value", "<?php $a = $this->ratio = null;", "<?php $this->ratio = null;\n$a = null;"},
		{"superglobal element", "<?php $a = $_GET['k'] = 1;", "<?php $_GET['k'] = 1;\n$a = 1;"},
		{"outer key effect", "<?php $x[$i++] = $y[$i++] = 0;", ""},
		{"outer property key", "<?php $a[$x->p] = $b = 5;", ""},
		{"outer element key", "<?php $a[$k['i']] = $b = 7;", ""},
		{"outer plain key", "<?php $c[$i] = $i = 9;", "<?php $i = 9;\n$c[$i] = 9;"},
		{"outer push", "<?php $c[] = $i = f();", "<?php $i = f();\n$c[] = $i;"},
		{"dynamic variable value", "<?php $a = $b = ${name()};", "<?php $b = ${name()};\n$a = $b;"},
		{"class expression constant", "<?php $c = $d = make()::N;", "<?php $d = make()::N;\n$c = $d;"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != 1 {
				t.Fatalf("findings: got %v, want 1", got)
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
