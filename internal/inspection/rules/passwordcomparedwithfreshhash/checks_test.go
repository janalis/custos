package passwordcomparedwithfreshhash

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PasswordComparedWithFreshHash"}})
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
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"$v=password_hash($p,PASSWORD_DEFAULT)+1;", 0},
		{"$v=password_hash()===$stored;", 0},
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
		{"namespace Site;function password_verify($p,$h){return false;}function f($password,$stored){return $password===password_hash($stored,PASSWORD_DEFAULT);}", "\\password_verify(", 1},
		{"function f($p,$stored){return $stored===password_hash($p,9999);}", "", 0},
		{"function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT,['cost'=>-1]);}", "", 0},
		{"function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT,[],1);}", "", 0},
		{"use function strlen as password_verify; function f($p,$stored){return $stored===password_hash($p,PASSWORD_DEFAULT);}", "\\password_verify(", 1},
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

// Before PHP 5.5 there is no password_verify() to suggest.
func TestBeforePasswordAPI(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP54, Only: []string{"PasswordComparedWithFreshHash"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Analyze(syntax.Parse("old.php", []byte("<?php $v=password_hash($p,PASSWORD_DEFAULT)===$saved;"), syntax.Options{})); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
