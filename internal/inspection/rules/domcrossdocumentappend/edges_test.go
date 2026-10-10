package domcrossdocumentappend

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
		{"$d=new DOMDocument(); $d->appendChild($d->createElement('x'));", 0},
		{"$d=new DOMDocument();$e=new DOMDocument();$d->insertBefore($e->createTextNode('t'));", 1},
		{"class DOMDocument {function appendChild($x){}} $d=new DOMDocument();$d->appendChild($other);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomCrossDocumentAppend"}})
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
