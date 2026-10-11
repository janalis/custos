package gmpautomaticbasechangesdecimalinput

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"gmp_init('024',0);", 1},
		{"gmp_init(num:'024');", 1},
		{"gmp_init('024',base:0);", 1},
		{"gmp_init('024',10);", 0},
		{"gmp_init('0007');", 0},
		{"gmp_init('0000');", 0},
		{"gmp_init('09');", 0},
		{"gmp_init('0x20');", 0},
		{"gmp_init('2');", 0},
		{"gmp_init($s);", 0},
		{"gmp_init('024',$base);", 0},
		{"gmp_init('024',1-1);", 0},
		{"gmp_init('024' /* retain */);", 1},
		{"strlen('024');", 0},
		{"const ZERO = 0; gmp_init('024', ZERO);", 1},
		{"gmp_init('024', ...$rest);", 0},
		{"gmp_init('024', bogus: 1);", 1},
		{"gmp_init(...['024']);", 0},
		{"function f(){ $raw = '024'; $copy = $raw; $raw++; return gmp_init($copy); }", 0},
		{"function f(){ $raw = '024'; $copy = $raw; return gmp_init($copy); }", 1},
		{"function f(){ $raw = '024'; $r = gmp_init($raw); try { g(); } catch (\\Exception $raw) {} return $r; }", 0},
		{"function f($o){ $raw = '024'; $r = gmp_init($raw); extract($o); return $r; }", 0},
		{"function f($q){ $raw = '024'; $r = gmp_init($raw); parse_str($q); return $r; }", 0},
		{"$raw = '024'; gmp_init($raw); function g(){ $GLOBALS['raw'] = '9'; }", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GmpAutomaticBaseChangesDecimalInput"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
