package dommissingattributecheckedasnull

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
		{"function f(DOMElement $e){$e->getAttribute('x') == null; $e->getAttribute('x') === '';}", 0},
		{"function f(DOMElement $e){$v=$e->getAttribute('x'); null !== $v;}", 1},
		{"function f(?DOMElement $e){$e?->getAttribute('x') === null;}", 0},
		{"function f(?DOMElement $e){$v=$e?->getAttribute('x'); $v === null;}", 0},
		{"class A{public DOMElement $el; public array $els; function m(): DOMElement {}}" +
			"function f(?A $a){$a?->el->getAttribute('x') === null; $a?->m()->getAttribute('x') === null; $a?->els[0]->getAttribute('x') !== null;}", 0},
		{"class A{public DOMElement $el; function m(): DOMElement {}}" +
			"function f(A $a){$a->el->getAttribute('x') === null; $a->m()->getAttribute('x') === null;}", 2},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomMissingAttributeCheckedAsNull"}})
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
