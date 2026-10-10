package contentdispositionfilenameneedsquoting

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
		{"header('Content-Disposition: attachment; filename=a b');", 1},
		{"strlen('x');", 0},
		{"header('Content-Disposition: attachment');", 0},
		{"header('Content-Disposition: attachment; filename=a; filename=b');", 0},
		{"header('Content-Disposition: attachment; filename=');", 0},
		{"header('Content-Disposition: attachment; filename=\"a b\"');", 0},
		{"header('Content-Disposition: attachment; filename=a.txt');", 0},
		{"$h='Content-Disposition: attachment; filename=a b';header($h);", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentDispositionFilenameNeedsQuoting"}})
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
