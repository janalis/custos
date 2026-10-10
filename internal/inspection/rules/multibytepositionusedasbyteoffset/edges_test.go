package multibytepositionusedasbyteoffset

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
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,\"UTF-8\");echo substr($s,$p,1);", 1, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"missing\",0,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"\",0,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,\"ISO-8859-1\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,$encoding);echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",2,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"test\";$p=mb_strpos($s,\"t\",0,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,\"UTF-8\");echo $xsubstr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,\"UTF-8\");echo substr(\"other\",$p,1);", 0, ""},
		{"$s=\"été\";$p=strpos($s,\"t\",0,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_stripos($s,\"t\",0,\"UTF-8\");echo substr($s,$p,1);", 1, ""},
		{"$s=\"été\";$p=mb_stripos($s,\"é\",0,\"UTF-8\");echo substr($s,$p,1);", 0, ""},
		{"$s=\"été\";$p=mb_strpos($s,\"t\",0,\"UTF-8\");echo substr($s,1,1);", 0, ""},
		{"echo substr($source,$p,1);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"MultibytePositionUsedAsByteOffset"}, PHP: phpversion.MustParse(tc.php)}
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
