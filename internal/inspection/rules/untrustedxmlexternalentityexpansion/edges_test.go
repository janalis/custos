package untrustedxmlexternalentityexpansion

import (
	"strings"
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
		{"$_POST[\"xml\"]=\"<root/>\";$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],8388614);", 1, "8.3"},
		{"if($x){}$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{"touch(\"file\");$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{"$x=1;$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{strings.Repeat("$d=new DOMDocument();", 65) + "$d->loadXML($_GET[\"xml\"],6);", 0, ""},
		{"$d=new DOMDocument(clearRequest());$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{"$_POST=new DOMDocument();$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],6);", 0, ""},
		{"other(); $x->other(); echo $x[0];", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($xml,LIBXML_NOENT|LIBXML_DTDLOAD);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($obj[\"xml\"],6);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($settings[\"xml\"],6);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($_GET[\"xml\"],$opts);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($_GET[\"xml\"],LIBXML_NOENT);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($_GET[\"xml\"],LIBXML_NOENT|LIBXML_DTDLOAD|LIBXML_NONET);", 1, ""},
		{"$d=new DOMDocument();$d->loadXML($_POST[\"xml\"],8388614);", 0, ""},
		{"$d=new DOMDocument();$d->loadXML($obj->data[\"xml\"],6);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"UntrustedXmlExternalEntityExpansion"}, PHP: phpversion.MustParse(tc.php)}
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
