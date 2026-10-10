package varyheaderoverwritesearlierdimensions

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
		{"strlen('x');", 0},
		{"header('Vary: Origin');", 0},
		{"header('Vary: Origin');header('Vary: *');", 0},
		{"header('Vary: Origin');header('Vary: Accept',false);", 0},
		{"header('Vary: Origin');header('Vary: Origin, Accept');", 0},
		{"header('Vary: Origin');header('Vary: bad token');", 0},
		{"header('Vary: bad token');header('Vary: Origin');", 0},
		{"header('Vary: ,');header('Vary: Origin');", 0},
		{"header('Vary: Origin');header('Vary: ,');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"VaryHeaderOverwritesEarlierDimensions"}})
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
