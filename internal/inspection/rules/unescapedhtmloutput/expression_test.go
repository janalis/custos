package unescapedhtmloutput

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestUnsafeExpressionRange(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"concatenation", `echo '<p>'.($_GET['first']).'</p>';`, `$_GET['first']`},
		{"comma operands", `echo '<p>', $_GET['first'], '</p>', $_GET['second'];`, `$_GET['first']|$_GET['second']`},
		{"print", `print '<p>'.$_POST['first'].'</p>';`, `$_POST['first']`},
		{"interpolation", `echo "<p>{$_GET['first']}</p>";`, `$_GET['first']`},
		{"unbraced interpolation", `echo "<p>$_GET[first]</p>";`, `$_GET[first]`},
		{"byte offsets", `echo "<p>é🙂{$_GET['first']}</p>";`, `$_GET['first']`},
		{"heredoc", "echo <<<HTML\n<p>{$_GET['first']}</p>\nHTML;", `$_GET['first']`},
		{"attribute", `echo '<p title="'.$_GET['first'].'">';`, `$_GET['first']`},
		{"transformed value", `echo '<p>'.strtoupper($_GET['first']);`, `strtoupper($_GET['first'])`},
		{"conditional value", `echo '<p>'.($_GET['flag'] ? $_GET['first'] : 'none');`, `$_GET['flag'] ? $_GET['first'] : 'none'`},
		{"escaped prefix", `echo '<p>'.htmlspecialchars($_GET['first']).$_GET['second'];`, `$_GET['second']`},
		// Each unsafe value is reported, read in the context the text intends.
		{"every unsafe value", `echo '<p>'.$_GET['first'].'<p>'.$_GET['second'];`, `$_GET['first']|$_GET['second']`},
		{"unsafe value then attribute", `echo '<p>'.$_GET['first'].'<a title="'.$_GET['second'].'">';`, `$_GET['first']|$_GET['second']`},
		{"opaque value ends proof", `echo '<p>'.$_GET['first'].$prefix.'<p>'.$_GET['second'];`, `$_GET['first']`},
		{"operand effect before output", `echo '<p>'.$_GET['first'].changeResponse();`, ""},
		{"later operand effect", `echo '<p>', $_GET['first'], changeResponse();`, `$_GET['first']`},
		{"unknown prefix", `echo $prefix.'<p>'.$_GET['first'];`, ""},
		{"inert response", `header('Content-Type: text/plain');echo '<p>'.$_GET['first'];`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("<?php " + tc.source)
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnescapedHtmlOutput"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("output.php", src, syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatal(f.Errors)
			}
			got := e.Analyze(f)
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected findings: %+v", got)
				}
				return
			}
			wants := strings.Split(tc.want, "|")
			if len(got) != len(wants) {
				t.Fatalf("got %d findings, want %d: %+v", len(got), len(wants), got)
			}
			for i, want := range wants {
				if actual := string(src[got[i].Span.Start:got[i].Span.End]); actual != want {
					t.Fatalf("highlighted %q, want %q", actual, want)
				}
			}
		})
	}
}
