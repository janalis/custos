package zipopentruthyresult

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
		{"$z=new ZipArchive(); $z->open($p);", 0, ""},
		{"$z=new ZipArchive(); while($z->open($p)) {}", 1, ""},
		{"$z=new ZipArchive(); if(!$z->open($p)) {}", 1, ""},
		{"$z=new ZipArchive(); $x=!$z->open($p);", 1, ""},
		{"$z=new ZipArchive(); if($z->open($p)===true) {}", 0, ""},
		{"$z=new ZipArchive(); echo strlen($z->open($p));", 0, ""},
		{"$z=new ZipArchive(); if($z->open($p)&&$ready) {}", 0, ""},
		{"$z=new ZipArchive(); if(1){$z->open($p);}", 0, ""},
		{"$z=new ZipArchive(); if((($z->open($p)))) {}", 1, ""},
		{"$z=new ZipArchive();echo +$z->open($p);", 0, ""},
		{"$z=new ZipArchive();if(!/* preserve */($z->open($p))){}", 1, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ZipOpenTruthyResult"}, PHP: phpversion.MustParse(tc.php)}
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
