package domappendmovesnodeunexpectedly

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
		{"$d=new DOMDocument();$a=$d->createElement('a');$b=$d->createElement('b');$n=$d->createElement('n');$a->appendChild($n);$a->removeChild($n);$b->appendChild($n);", 0},
		{"$d=new DOMDocument();$a=$d->createElement('a');$n=$d->createElement('n');$a->appendChild($n);$a->appendChild($n);", 0},
		{"$d=new DOMDocument();$d->appendChild($unknown);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomAppendMovesNodeUnexpectedly"}})
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
