package intlformattingfailureunchecked

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
		{"echo strlen($s);", 0, ""},
		{"$f=new IntlDateFormatter(\"en\",1,1);echo strlen($f->format($v));", 1, ""},
		{"$f=new NumberFormatter(\"en\",1);$s=$f->format($v);if($s===false){return;}echo strlen($s);", 0, ""},
		{"$f=new NumberFormatter(\"en\",1);$s=$f->format($v);$s=\"ok\";echo strlen($s);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"IntlFormattingFailureUnchecked"}, PHP: phpversion.MustParse(tc.php)}
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
