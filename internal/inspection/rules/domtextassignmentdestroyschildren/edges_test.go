package domtextassignmentdestroyschildren

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
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createElement('y'));$n->textContent='new';`, 1},
		{`$d=new DOMDocument();$n=$d->createElement('x');$n->appendChild($d->createTextNode('y'));$n->textContent='new';`, 0},
		{`$d=new DOMDocument();$n=$d->createElement('x');$child=$d->createElement('y');$n->appendChild($child);$n->removeChild($child);$n->textContent='new';`, 0},

		{"$d=new DOMDocument();$d->loadXML('<list><item/></list>');$d->documentElement->other='text';", 0},
		{"$d=new DOMDocument();$d->loadXML('<list/>');$d->documentElement->nodeValue='text';", 0},
		{"$d=new DOMDocument();$d->loadXML('<list><item/></list>');$d->documentElement->textContent='text';", 1},
		{"$d=new DOMDocument();$d->loadXML('<list><item/></list>');mutate($d);$d->documentElement->textContent='text';", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomTextAssignmentDestroysChildren"}})
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
