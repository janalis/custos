package domlivenodelistremovalskipsnodes

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, src := range []string{
		`function f(DOMDocument $d,DOMNodeList $nodes){$d->removeChild($nodes->item($i));}`,
		`function f(DOMXPath $x){$nodes=$x->query("*");if($nodes!==false){for($i=0;$i<$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}}`,
		`function f(DOMNodeList $nodes){for($i=0;$i<$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}`,
		`function f(DOMDocument $d){$nodes=$d->getElementsByTagName("x");$i=0;$n=$nodes->item($i);for($i=0;$i<$nodes->length;$i++) $n->parentNode->removeChild($n);}`,
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');$d->removeChild($unknown);}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');$n=$nodes->item(0);$n->parentNode->removeChild($n);}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');$n=$nodes->item($i);$d->removeChild($n);}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');$n=$nodes->item($i);$n->ownerDocument->removeChild($n);}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for(;$i<$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for(0;$i<$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=2;$i<$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i>$nodes->length;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<3;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->other;$i++){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->length;$i--){$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->length;$i++)$nodes->item($i)->parentNode->removeChild($nodes->item($i));}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->length;$i++){if($flag){} $n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->length;$i++){$unrelated=1;$n=$nodes->item($i);$n->parentNode->removeChild($n);}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');for($i=0;$i<$nodes->length;$i++){$n=$nodes->item($i);if($flag){$n->parentNode->removeChild($n);}}}",
		"function f(DOMDocument $d){$nodes=$d->getElementsByTagName('x');$n=$nodes->item($i);$n->parentNode->removeChild($n);}",
		"function f(DOMNodeList $nodes){$n=$nodes->item($i);$n->parentNode->removeChild($n);}",
	} {
		t.Run(src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DomLiveNodeListRemovalSkipsNodes"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edges.php", []byte("<?php "+src), syntax.Options{}))
			if len(got) != 0 {
				t.Fatalf("unexpected: %+v", got)
			}
		})
	}
}
