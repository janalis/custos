package domlivenodelistremovalskipsnodes

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const head = "<?php function f(DOMDocument $d, $o) { $nodes = $d->getElementsByTagName('p'); for ($i = 0; $i < $nodes->length; $i++) { "
	const bound = "$n = $nodes->item($i); $n->parentNode->removeChild($n); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"plain", bound, 1},
		{"unrelated trailing statement", bound + "$o->done(); if ($o) { $o->log($n); }", 1},
		{"post decrement", bound + "$i--;", 0},
		{"pre decrement", bound + "--$i;", 0},
		{"compound correction", bound + "$i -= 1;", 0},
		{"assigned correction", bound + "$i = $i - 1;", 0},
		{"nested correction", bound + "if ($o) { $i--; }", 0},
		{"index passed on", bound + "$o->adjust($i);", 0},
		{"variable variable", bound + "$$o = -1;", 0},
		{"break", bound + "break;", 0},
		{"continue", bound + "if ($o) { continue; }", 0},
		{"return", bound + "return;", 0},
		{"throw", bound + "throw new RuntimeException('x');", 0},
		{"exit", bound + "exit(1);", 0},
		{"goto", bound + "goto done;", 0},
		{"direct item", "$nodes->item($i)->parentNode->removeChild($nodes->item($i));", 1},
		{"direct item spacing", "$nodes->item( $i )->parentNode->removeChild($nodes->item($i));", 1},
		{"direct item corrected", "$nodes->item($i)->parentNode->removeChild($nodes->item($i)); $i--;", 0},
		{"direct other index", "$nodes->item(0)->parentNode->removeChild($nodes->item($i));", 0},
		{"direct other list", "$o->item($i)->parentNode->removeChild($nodes->item($i));", 0},
		{"bound child direct receiver", "$n = $nodes->item($i); $nodes->item($i)->parentNode->removeChild($n);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", []byte(head+tc.body+"} done: }"), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
