package comparisonoperandsorder

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
		php          phpversion.Version
		style        analysis.ComparisonStyle
		fixed        string
	}{
		{"keyword before", "<?php function t() { return'a'==foo(); }", phpversion.PHP85, analysis.StyleRegular, "<?php function t() { return foo()=='a'; }"},
		{"qualified name", "<?php function t() { return'a'==\\foo(); }", phpversion.PHP85, analysis.StyleRegular, "<?php function t() { return \\foo()=='a'; }"},
		{"echo before", "<?php echo'a'==foo();", phpversion.PHP85, analysis.StyleRegular, "<?php echo foo()=='a';"},
		{"quote after keyword", "<?php function t($x) { return$x=='a'; }", phpversion.PHP85, analysis.StyleYoda, "<?php function t($x) { return'a'==$x; }"},
		{"constant after keyword", "<?php function t($x) { return$x==null; }", phpversion.PHP85, analysis.StyleYoda, "<?php function t($x) { return null==$x; }"},
		{"keyword after", "<?php $r = null==$a[0]and$b;", phpversion.PHP85, analysis.StyleRegular, "<?php $r = $a[0]==null and$b;"},
		{"variable before keyword", "<?php $r = $x=='a'and $b;", phpversion.PHP85, analysis.StyleYoda, "<?php $r = 'a'==$x and $b;"},
		{"interpolation vs increment", "<?php $r = \"v{$i}\" == 'v' . $i++;", phpversion.PHP85, analysis.StyleRegular, ""},
		{"increment vs interpolation", "<?php $r = $i++ == \"$i\";", phpversion.PHP85, analysis.StyleYoda, ""},
		{"interpolated call", "<?php $r = \"{$o->next()}\" == $o->pos;", phpversion.PHP85, analysis.StyleRegular, ""},
		{"pure interpolation", "<?php $r = \"v{$i}\" == $j;", phpversion.PHP85, analysis.StyleRegular, "<?php $r = $j == \"v{$i}\";"},
		{"literal vs call", "<?php $r = 'v' == next($i);", phpversion.PHP85, analysis.StyleRegular, "<?php $r = next($i) == 'v';"},
		{"heredoc 7.2", "<?php\nvar_dump(<<<EOT\nx\nEOT\n === $a);\n", phpversion.PHP72, analysis.StyleRegular, ""},
		{"heredoc 7.2 yoda", "<?php\n$r = $a === <<<EOT\nx\nEOT;\n", phpversion.PHP72, analysis.StyleYoda, ""},
		{"nested heredoc 7.2", "<?php\n$r = null === $a . <<<'EOT'\nx\nEOT;\n", phpversion.PHP72, analysis.StyleRegular, ""},
		{"heredoc 7.3", "<?php\n$r = $a === <<<EOT\nx\nEOT;\n", phpversion.PHP73, analysis.StyleYoda, "<?php\n$r = <<<EOT\nx\nEOT === $a;\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.php, EnableAll: true, ComparisonStyle: tc.style})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte(tc.source)
			opt := syntax.Options{Version: tc.php}
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
