package domcrossdocumentappend

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewAdoptedNodeOwnership(t *testing.T) {
	const docs = "$a = new DOMDocument(); $b = new DOMDocument(); $c = new DOMDocument(); "
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"adopted by destination", "$n = $b->createElement('x'); $a->adoptNode($n); $a->appendChild($n);", 0},
		{"adopted with named argument", "$n = $b->createElement('x'); $a->adoptNode(node: $n); $a->appendChild($n);", 0},
		{"adoption result assigned", "$n = $b->createElement('x'); $m = $a->adoptNode($n); $a->appendChild($n);", 0},
		{"adopted through alias", "$n = $b->createElement('x'); $m = $n; $a->adoptNode($m); $a->appendChild($n);", 0},
		{"adopted receiver", "$n = $b->createElement('x'); $a->adoptNode($n); $n->appendChild($a->createElement('y'));", 0},
		{"adopted receiver keeps foreign child", "$n = $b->createElement('x'); $a->adoptNode($n); $n->appendChild($b->createElement('y'));", 1},
		{"adopted by third document", "$n = $b->createElement('x'); $c->adoptNode($n); $a->appendChild($n);", 1},
		{"adopted away from destination", "$n = $a->createElement('x'); $b->adoptNode($n); $a->appendChild($n);", 1},
		{"adopted back to source", "$n = $b->createElement('x'); $a->adoptNode($n); $b->adoptNode($n); $a->appendChild($n);", 1},
		{"adoption result of third document", "$n = $b->createElement('x'); $m = $c->adoptNode($n); $a->appendChild($m);", 1},
		{"adoption after append", "$n = $b->createElement('x'); $a->appendChild($n); $a->adoptNode($n);", 1},
		{"import does not move", "$n = $b->createElement('x'); $a->importNode($n); $a->appendChild($n);", 1},
		{"no adoption", "$n = $b->createElement('x'); $a->appendChild($n);", 1},
		{"conditional adoption", "$n = $b->createElement('x'); if ($k) { $a->adoptNode($n); } $a->appendChild($n);", 0},
		{"adoption in condition", "$n = $b->createElement('x'); if ($a->adoptNode($n)) { echo 1; } $a->appendChild($n);", 0},
		{"adopted ancestor", "$n = $b->createElement('x'); $p = $b->createElement('p'); $p->appendChild($n); $a->adoptNode($p); $a->appendChild($n);", 0},
		{"adopted by unknown document", "$n = $b->createElement('x'); $q->adoptNode($n); $a->appendChild($n);", 0},
		{"adopted by parameter document", "$n = $b->createElement('x'); $d->adoptNode($n); $a->appendChild($n);", 0},
		{"inline node after unrelated adoption", "$m = $b->createElement('m'); $a->adoptNode($m); $a->appendChild($b->createElement('z'));", 1},
		{"document variables after adoption", "$m = $b->createElement('m'); $a->adoptNode($m); $d2 = $a; $d2->appendChild($b->createElement('z'));", 1},
		{"unrelated call before append", "$n = $b->createElement('x'); $a->saveXML(); strlen('x'); $a->appendChild($n);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomCrossDocumentAppend"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php function f(DOMDocument $d, $k) { "+docs+tc.src+" }"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
