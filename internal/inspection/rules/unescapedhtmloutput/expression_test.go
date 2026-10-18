package unescapedhtmloutput

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestUnsafeExpressionRange(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"concatenation", `echo '<p>'.($_GET['first']).'</p>';`, `$_GET['first']`},
		{"comma operands", `echo '<p>', $_GET['first'], '</p>', $_GET['second'];`, `$_GET['first']`},
		{"print", `print '<p>'.$_POST['first'].'</p>';`, `$_POST['first']`},
		{"interpolation", `echo "<p>{$_GET['first']}</p>";`, `$_GET['first']`},
		{"unbraced interpolation", `echo "<p>$_GET[first]</p>";`, `$_GET[first]`},
		{"byte offsets", `echo "<p>é🙂{$_GET['first']}</p>";`, `$_GET['first']`},
		{"heredoc", "echo <<<HTML\n<p>{$_GET['first']}</p>\nHTML;", `$_GET['first']`},
		{"attribute", `echo '<p title="'.$_GET['first'].'">';`, `$_GET['first']`},
		{"transformed value", `echo '<p>'.strtoupper($_GET['first']);`, `strtoupper($_GET['first'])`},
		{"conditional value", `echo '<p>'.($_GET['flag'] ? $_GET['first'] : 'none');`, `$_GET['flag'] ? $_GET['first'] : 'none'`},
		{"escaped prefix", `echo '<p>'.htmlspecialchars($_GET['first']).$_GET['second'];`, `$_GET['second']`},
		{"unsafe value ends proof", `echo '<p>'.$_GET['first'].'<p>'.$_GET['second'];`, `$_GET['first']`},
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
			if len(got) != 1 {
				t.Fatalf("got %d findings, want 1: %+v", len(got), got)
			}
			if actual := string(src[got[0].Span.Start:got[0].Span.End]); actual != tc.want {
				t.Fatalf("highlighted %q, want %q", actual, tc.want)
			}
		})
	}
}
