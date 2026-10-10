package zipclosefailureunchecked

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
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}$z->addFile($p);$z->close();return true;}", 1, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)===true){return false;}$z->addFile($p);$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}$z->setArchiveComment(\"x\");$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){echo \"failed\";}$z->addFile($p);$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}$z->addFile($p);return $z->close();}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}if($ok){}$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();$z->addFile($p);$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}$z->addFile($p);$z->close();return false;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==true){return false;}$other=new ZipArchive();$other->addFile($p);$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->open($p)!==false){return false;}$z->addFile($p);$z->close();return true;}", 0, ""},
		{"function save($p){$z=new ZipArchive();if($z->getNameIndex(0)!==true){return false;}$z->addFile($p);$z->close();return true;}", 0, ""},
		{"function f(ZipArchive $z){$z->close();return true;}", 0, ""},
		{"function f(ZipArchive $z){$x=1;$x=2;$z->close();return true;}", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ZipCloseFailureUnchecked"}, PHP: phpversion.MustParse(tc.php)}
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
