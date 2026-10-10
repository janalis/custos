package libxmlerrorbuffernevercleared

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
		{"foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"libxml_use_internal_errors(false);foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){simplexml_load_string($xml);}", 1, ""},
		{"libxml_use_internal_errors(true);for($i=0;$i<2;$i++){simplexml_load_string($xml);}", 1, ""},
		{"libxml_use_internal_errors(true);while($ok){simplexml_load_string($xml);}", 1, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml) simplexml_load_string($xml);", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){other();}", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){$x=1;}", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){if($ok){} }", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){}", 0, ""},
		{"libxml_use_internal_errors(true);foreach($docs as $xml){simplexml_load_string($xml);}libxml_clear_errors();", 0, ""},
		{"libxml_use_internal_errors(true);$x=1;foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"libxml_use_internal_errors(true);$x=new stdClass();foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"if($ok){}foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"libxml_use_internal_errors(true);other();foreach($docs as $xml){simplexml_load_string($xml);}", 0, ""},
		{"libxml_use_internal_errors(true);$d=new DOMDocument();foreach($docs as $xml){$d->other();}", 0, ""},
		{"$d=new DOMDocument();foreach($docs as $xml){$d->loadXML($xml);}", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"LibxmlErrorBufferNeverCleared"}, PHP: phpversion.MustParse(tc.php)}
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
