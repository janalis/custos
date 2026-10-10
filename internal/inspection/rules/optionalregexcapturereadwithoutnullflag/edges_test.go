package optionalregexcapturereadwithoutnullflag

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
		{"echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo $m[1];", 1, ""},
		{"if($ok){}echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$other);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo isset($m[1]);", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo ($m[1]??\"\");", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo $m[$index];", 0, ""},
		{"preg_match($pattern,\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/[/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/i\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/(?:(a))?b/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",$subject,$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"é\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\\n\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/a{4,2}/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/\\\\b(a)?b/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo $m[0];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m);echo $m[9];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"ab\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"none\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?b$/\",\"b\",$m,$flags);echo $m[1];", 0, ""},
		{"preg_match(\"/^(a)?(b)$/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/(a)/\",\"a\",$obj->m);echo $obj->m[1];", 0, ""},
		{"$m=[];echo $m[1];", 0, ""},
		{"preg_match(\"/(a)?/\",\"b\",$m);echo $m[1];", 0, ""},
		{"preg_match(\"/(a)?/\",str_repeat(\"a\",1024),$m);echo $m[1];", 0, ""},
		{"preg_match(\"/(a)/\",\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",$m);echo $m[1];", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"OptionalRegexCaptureReadWithoutNullFlag"}, PHP: phpversion.MustParse(tc.php)}
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
