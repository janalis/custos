package contentlengthusescharactercount

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthUsesCharacterCount"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f($body){ini_set('zlib.output_compression',true);header('Content-Length: '.mb_strlen($body));echo $body;}", 0},
		{"header('Content-Length: '.mb_strlen('é'));echo 'é';", 1},
		{"header('Content-Length: '.mb_strlen(1));echo 1;", 0},
		{"header('Content-Length: '.mb_strlen(1.5));echo 1.5;", 0},
		{"header('Content-Length: '.mb_strlen(true));echo true;", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));echo $body;}", 1},
		{"function f($body){header('Content-Length: '.grapheme_strlen($body));echo $body;}", 1},
		{"function f($body){header('Content-Length: '.mb_strlen($body));echo 'different';}", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));$body='replacement';echo $body;}", 0},
		{"function f($body){ob_start('compress');header('Content-Length: '.mb_strlen($body));echo $body;}", 0},
		{"header('Content-Length: '.mb_strlen('ascii'));echo 'ascii';", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));echo $body,'tail';}", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));echo $body;echo 'tail';}", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));echo $body;transformOutput();}", 0},
		{"function f($body){header('Content-Length: '.mb_strlen($body));print $body;}", 0},
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"header('Content-Length: 2'); header($prefix.mb_strlen($body));", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// TestFixes checks the quick-fix of the one finding on short sources: how
// many it has and, where there is one, that its single edit inserts needle
// and leaves code that parses.
func TestFixes(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src, needle string
		fixes       int
	}{
		{"namespace Site;function strlen($s){return 0;}function f($body){header('Content-Length: '.mb_strlen($body));echo $body;}", "\\strlen($body)", 1},
		{"function f($body){header('Content-Length: '.mb_strlen($body,'invalid'));echo $body;}", "", 0},
		{"function f($body,$encoding){header('Content-Length: '.mb_strlen($body,$encoding));echo $body;}", "", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			src := "<?php " + tc.src
			got := e.Analyze(syntax.Parse("fix.php", []byte(src), syntax.Options{}))
			if len(got) != 1 || len(got[0].Fixes) != tc.fixes {
				t.Fatalf("got %+v, want one with %d fixes", got, tc.fixes)
			}
			if tc.fixes == 0 {
				return
			}
			edits := got[0].Fixes[0].Edits()
			if len(edits) != 1 || !strings.Contains(edits[0].NewText, tc.needle) {
				t.Fatalf("edits %+v, want one inserting %q", edits, tc.needle)
			}
			fixed := src[:edits[0].Span.Start] + edits[0].NewText + src[edits[0].Span.End:]
			if f := syntax.Parse("fixed.php", []byte(fixed), syntax.Options{}); len(f.Errors) > 0 {
				t.Fatalf("fix does not parse: %s: %+v", fixed, f.Errors)
			}
		})
	}
}
