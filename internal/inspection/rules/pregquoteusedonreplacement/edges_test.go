package pregquoteusedonreplacement

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
		{"preg_replace(\"/x/\",$replacement,$s);", 0, ""},
		{"preg_replace(\"/x/\",preg_quote($r),$s);", 0, ""},
		{"preg_replace(\"/x/\",preg_quote(\"$1\"),$s);", 0, ""},
		{"preg_replace(\"/x/\",preg_quote(\"plain\"),$s);", 0, ""},
		{"preg_replace(\"/x/\",preg_quote(\"a.b\",$delimiter),$s);", 1, ""},
		{"preg_replace(\"/x/\",preg_quote(/* keep */\"a.b\"),$s);", 1, ""},
		{"preg_replace(\"/x/\",preg_quote(\"a.b\"),$s);", 1, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"PregQuoteUsedOnReplacement"}, PHP: phpversion.MustParse(tc.php)}
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
