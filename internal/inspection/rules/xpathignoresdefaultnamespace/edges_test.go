package xpathignoresdefaultnamespace

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
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$x->query(\"//item\");", 1, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$x->query(\"/item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$x->query(\"//item[1]\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root ><item/></root>');$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item xmlns=\"\"/></root>');$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><other/></root>');$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/>');$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$x->query($query);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new stdClass();$x->query(\"//item\");", 0, ""},
		{"function f(DOMXPath $x){$x->query(\"//item\");}", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($xml);$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$d->loadXML(\"<root><item/></root>\");$x->query(\"//item\");", 0, ""},
		{"$d=new DOMDocument();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);if($ok){}$x->query(\"//item\");", 0, ""},
		{"(new DOMXPath($d))->query(\"//item\");", 0, ""},
		{"class Doc extends DOMDocument {public function loadXML(string $s,int $options=0):bool{return true;}}$d=new Doc();$d->loadXML('<root xmlns=\"urn:a\"><item/></root>');$x=new DOMXPath($d);$x->query(\"//item\");", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"XPathIgnoresDefaultNamespace"}, PHP: phpversion.MustParse(tc.php)}
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
