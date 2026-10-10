package codepointslicesplitsgrapheme

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
		{"mb_substr(\"e\\u{0301}\",0,1,\"UTF-8\");", 1, ""},
		{"mb_substr(\"e\\u{0301}\",1,1,\"UTF-8\");", 1, ""},
		{"mb_substr(\"e\\u{0301}\",-1,1,\"UTF-8\");", 1, ""},
		{"mb_substr(\"e\\u{0301}\",-9,1,\"UTF-8\");", 1, ""},
		{"mb_substr(\"e\\u{0301}\",9,1,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",0,-1,\"UTF-8\");", 1, ""},
		{"mb_substr(\"e\\u{0301}\",0,9,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",0,0,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",0,$n,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",$p,1,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",0,1,\"ISO-8859-1\");", 0, ""},
		{"mb_substr($s,0,1,\"UTF-8\");", 0, ""},
		{"mb_substr(\"e\\u{0301}\",0,2,\"UTF-8\");", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"CodePointSliceSplitsGrapheme"}, PHP: phpversion.MustParse(tc.php)}
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
