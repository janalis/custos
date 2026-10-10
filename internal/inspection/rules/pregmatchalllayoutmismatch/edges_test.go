package pregmatchalllayoutmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
		php    string
	}{
		{"other(); $x->other(); echo $x[0];", 0, ""},
		{"echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo $m[2][0];", 1, ""},
		{"if($ok){}echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$other);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo isset($m[2][0]);", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo ($m[2][0]??\"\");", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo $m[$index][0];", 0, ""},
		{"preg_match_all($pattern,\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/[/\",\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/i\",\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/(?:(a))?b/\",\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",$subject,$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"é\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\\n\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/a{4,2}/\",\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/\\\\b(a)?b/\",\"b\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo $m[2];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo $m[2][-$n];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m);echo $m[-1][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m,256);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/^(a)?b$/\",\"b\",$m,$flags);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/(a)/\",\"a\",$obj->m);echo $obj->m[2][0];", 0, ""},
		{"$m=[];echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/(a)?/\",\"a\",$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/(a)/\",str_repeat(\"a\",1024),$m);echo $m[2][0];", 0, ""},
		{"preg_match_all(\"/(a)/\",\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",$m);echo $m[2][0];", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"PregMatchAllLayoutMismatch"}, PHP: phpversion.MustParse(tc.php)}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, cfg)
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
