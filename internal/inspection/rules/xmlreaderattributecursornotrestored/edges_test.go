package xmlreaderattributecursornotrestored

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f(XMLReader $r){if($r->moveToAttribute('x')){$r->nodeType == XMLReader::ELEMENT; $r->other===XMLReader::ELEMENT;$r->nodeType===DOMNode::ELEMENT_NODE;$r->nodeType===XMLReader::ATTRIBUTE;}}", 0},
		{"function f(XMLReader $r,XMLReader $other){if($r->moveToAttribute('x')){$other->nodeType===XMLReader::ELEMENT;}}", 0},
		{"function f(XMLReader $r){if($flag){$r->nodeType===XMLReader::ELEMENT;}}", 0},
		{"function f(XMLReader $r){if($r->moveToFirstAttribute()){XMLReader::ELEMENT===$r->nodeType;}}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"XmlReaderAttributeCursorNotRestored"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
