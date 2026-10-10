package iteratormaterializationkeycollision

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"iterator_to_array(missing());", 0},
		{"function values(){yield from other();} iterator_to_array(values());", 0},
		{"iterator_to_array($unknown);", 0},
		{"function rows(){yield 'a'=>1;yield 'b'=>2;} iterator_to_array(rows());", 0},
		{"function rows(){yield $unknown=>1;} iterator_to_array(rows());", 0},
		{"function rows(){yield 'a'=>1;yield 'a'=>2;} iterator_to_array(rows(),$unknown);", 0},
		{"function rows(){echo 'x';yield 'a'=>1;yield 'a'=>2;} iterator_to_array(rows());", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IteratorMaterializationKeyCollision"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
