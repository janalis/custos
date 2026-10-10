package zipentrystreamusedafterarchiveclose

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
		{"echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");$z->close();echo fread($h,1);", 1, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");other();echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");$z->close();echo fread($other,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");$other=new ZipArchive();$other->close();echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=&$z->getStream(\"x\");$z->close();echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");if($ok){}echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getNameIndex(0);$z->close();echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$x=1;$z->close();echo fread($h,1);", 0, ""},
		{"$z=new ZipArchive();$h=$z->getStream(\"x\");$z->close();echo fread($h->stream,1);", 0, ""},
		{"$z=new ZipArchive();if($ok){}$z->close();echo fread($h,1);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ZipEntryStreamUsedAfterArchiveClose"}, PHP: phpversion.MustParse(tc.php)}
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
