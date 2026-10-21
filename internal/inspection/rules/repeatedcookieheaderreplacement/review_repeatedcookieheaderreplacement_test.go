package repeatedcookieheaderreplacement

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewRepeatedCookieHeaderReplacement(t *testing.T) {
	const first = "header('Set-Cookie: a=1');"
	for _, tc := range []struct {
		name, src   string
		want, fixes int
	}{
		{"distinct names", first + "header('Set-Cookie: b=2');", 1, 1},
		{"distinct names with attributes", "header('Set-Cookie: a=1; Path=/');header('Set-Cookie: b=2; Path=/');", 1, 1},
		{"explicit true keeps report without fix", first + "header('Set-Cookie: b=2', true);", 1, 0},
		{"same name new value", "header('Set-Cookie: sid=abc; Path=/');header('Set-Cookie: sid=deleted; Max-Age=0; Path=/');", 0, 0},
		{"same name spacing and field case", first + "header('set-cookie:a = 2');", 0, 0},
		{"names differ by case", first + "header('Set-Cookie: A=2');", 1, 1},
		{"second line without pair", first + "header('Set-Cookie: flag');", 0, 0},
		{"second line pair only in attribute", first + "header('Set-Cookie: flag; Path=/x=y');", 0, 0},
		{"first line without pair", "header('Set-Cookie: flag');header('Set-Cookie: b=2');", 0, 0},
		{"empty name", first + "header('Set-Cookie: =2');", 0, 0},
		{"remove cookie headers", first + "header_remove('Set-Cookie');header('Set-Cookie: b=2');", 0, 0},
		{"remove cookie headers any case", first + "header_remove(' set-cookie ');header('Set-Cookie: b=2');", 0, 0},
		{"remove named argument", first + "header_remove(name: 'Set-Cookie');header('Set-Cookie: b=2');", 0, 0},
		{"remove every header", first + "header_remove();header('Set-Cookie: b=2');", 0, 0},
		{"remove unknown header", first + "header_remove($field);header('Set-Cookie: b=2');", 0, 0},
		{"remove on one branch", first + "if($reset){header_remove('Set-Cookie');}header('Set-Cookie: b=2');", 0, 0},
		{"remove other header", first + "header_remove('X-Powered-By');header('Set-Cookie: b=2');", 1, 1},
		{"remove before first header", "header_remove();" + first + "header('Set-Cookie: b=2');", 1, 1},
		{"remove after second header", first + "header('Set-Cookie: b=2');header_remove();", 1, 1},
		{"loop iterations are not paired", "foreach($xs as $x){header('Set-Cookie: b=2');header('Set-Cookie: a=1',false);}", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"RepeatedCookieHeaderReplacement"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php function f($xs,$reset,$field){"+tc.src+"}"), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
			if len(got) > 0 && len(got[0].Fixes) != tc.fixes {
				t.Fatalf("got %d fixes, want %d", len(got[0].Fixes), tc.fixes)
			}
		})
	}
}
