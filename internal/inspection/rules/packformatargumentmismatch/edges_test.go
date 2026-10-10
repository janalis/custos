package packformatargumentmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"pack('a10');", 1},
		{"pack('x2N');", 1},
		{"pack('N*',1,2);", 0},
		{"pack('N*N',1,2);", 1},
		{"pack('a*N', 'a');", 1},
		{"pack('N0');", 0},
		{"pack('N2',1,2,3);", 0},
		{"pack('x*');", 0},
		{"pack('!');", 0},
		{"pack('N999999999999999999999999999');", 0},
		{"pack('N1000001');", 0},
		{"pack($format,1);", 0},
		{"pack('N',...$args);", 0},
		{"pack(format:'N2',args:1);", 0},
		{"strlen('N2');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PackFormatArgumentMismatch"}})
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
